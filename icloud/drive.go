package icloud

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ivandeex/go-icloud/icloud/api"
)

// DriveService describes the Drive iCloud service
type DriveService struct {
	c       *Client
	svcRoot string
	docRoot string
	root    *DriveNode
}

// NewDrive returns new Drive service
func NewDrive(c *Client) (d *DriveService, err error) {
	d = &DriveService{c: c}
	if d.svcRoot, err = c.getWebserviceURL("drivews"); err != nil {
		return nil, err
	}
	if d.docRoot, err = c.getWebserviceURL("docws"); err != nil {
		return nil, err
	}
	return d, nil
}

// Root returns root folder
func (d *DriveService) Root() (*DriveNode, error) {
	root := d.root
	if root == nil {
		item, err := d.getNodeData("root")
		if err != nil {
			return nil, err
		}
		root = &DriveNode{
			d:     d,
			i:     item,
			ready: true,
		}
		d.root = root
	}
	return root, nil
}

// getNodeData returns node data
func (d *DriveService) getNodeData(nodeID string) (*api.DriveItem, error) {
	folder := dict{
		"drivewsid":   "FOLDER::com.apple.CloudDocs::" + nodeID,
		"partialData": false,
	}
	var res []api.DriveItem
	if err := d.c.post(d.svcRoot+"/retrieveItemDetailsInFolders", []dict{folder}, nil, &res); err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, errors.New("invalid node data")
	}
	return &res[0], nil
}

// DriveNode ...
type DriveNode struct {
	d     *DriveService
	i     *api.DriveItem
	ready bool
}

// Name of node
func (n *DriveNode) Name() string {
	name, ext := n.i.Name, n.i.Ext
	if name != "" && ext != "" {
		name += "." + ext
	}
	return name
}

// Size of node
func (n *DriveNode) Size() int64 {
	if n.i.Size == nil {
		return -1
	}
	return *n.i.Size
}

// Type of node
func (n *DriveNode) Type() string { return strings.ToLower(n.i.Type) }

// IsDir returns true if node is a folder, false if it's a file
func (n *DriveNode) IsDir() bool { return n.Type() == "folder" }

// Changed time of node
func (n *DriveNode) Changed() time.Time { return n.i.Changed }

// Modified time of node
func (n *DriveNode) Modified() time.Time { return n.i.Modified }

// LastOpened time of node
func (n *DriveNode) LastOpened() time.Time { return n.i.LastOpened }

// Children of node
func (n *DriveNode) Children() ([]*DriveNode, error) {
	if !n.IsDir() {
		return nil, errors.New("must be a folder to list children")
	}
	if !n.ready {
		item, err := n.d.getNodeData(n.i.DocID)
		if err != nil {
			return nil, err
		}
		n.i.Items = item.Items
		n.ready = true
	}
	children := []*DriveNode{}
	for _, item := range n.i.Items {
		children = append(children, &DriveNode{
			d: n.d,
			i: item,
		})
	}
	return children, nil
}

func (n *DriveNode) Dir() ([]string, error) {
	children, err := n.Children()
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, child := range children {
		names = append(names, child.Name())
	}
	return names, nil
}

func (n *DriveNode) Get(name string) (*DriveNode, error) {
	children, err := n.Children()
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		if child.Name() == name {
			return child, nil
		}
	}
	return nil, errors.New("child not found")
}

// Open file for reading
func (n *DriveNode) Open() (io.ReadCloser, error) {
	if n.IsDir() {
		return nil, errors.New("cannot open folder")
	}
	if n.Size() <= 0 {
		// iCloud returns 400 Bad Request for empty files
		return io.NopCloser(&bytes.Buffer{}), nil
	}
	return n.d.getFile(n.i.DocID)
}

// getFile returns an iCloud Drive file
func (d *DriveService) getFile(fileID string) (io.ReadCloser, error) {
	var docResult *api.DriveDocResult
	docURL := d.docRoot + "/ws/com.apple.CloudDocs/download/by_id?document_id=" + fileID
	if err := d.c.get(docURL, &docResult); err != nil {
		return nil, fmt.Errorf("cannot get download url for id %q: %w", fileID, err)
	}
	var fileURL string
	if docResult != nil {
		fileURL = docResult.DataToken.URL
	}
	if fileURL == "" {
		return nil, errors.New("failed to get file url")
	}
	var stream io.ReadCloser
	if err := d.c.get(fileURL, &stream); err != nil {
		return nil, fmt.Errorf("failed to download from id %q: %w", fileID, err)
	}
	return stream, nil
}

// Download node into local file
func (n *DriveNode) Download(path string) error {
	in, err := n.Open()
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("%s: cannot create file: %w", path, err)
	}
	_, err = io.Copy(out, in)
	errClose := out.Close()
	if err == nil {
		err = errClose
	}
	return err
}

// Upload new file to a folder
func (n *DriveNode) Upload(path string) error {
	f, err := os.Open(path)
	var fi os.FileInfo
	if err == nil {
		fi, err = f.Stat()
	}
	if err == nil {
		err = n.PutStream(f, path, fi.Size(), fi.ModTime())
	}
	return err
}

// PutStream uploads a file stream to a folder
func (n *DriveNode) PutStream(in io.Reader, path string, size int64, mtime time.Time) error {
	if !n.IsDir() {
		return errors.New("can only upload to a folder")
	}
	_, err := n.d.sendFile(n.i.DocID, in, path, size, mtime)
	return err
}

// sendFile sends new file to iCloud Drive
func (d *DriveService) sendFile(folderID string, in io.Reader, path string, size int64, mtime time.Time) (dict, error) {
	name := filepath.Base(path)
	mimeType := mime.TypeByExtension(filepath.Ext(name))
	docID, contentURL, err := d.getUploadContentWsURL(name, mimeType, size)
	if err != nil {
		return nil, err
	}

	// Prepare multipart body
	body := &bytes.Buffer{}
	mpWriter := multipart.NewWriter(body)
	partWriter, err := mpWriter.CreateFormFile(name, name)
	if err == nil {
		_, err = io.Copy(partWriter, in)
	}
	if errClose := mpWriter.Close(); err == nil {
		err = errClose
	}
	if closer, canClose := in.(io.ReadCloser); canClose {
		if errClose := closer.Close(); err == nil {
			err = errClose
		}
	}
	if err != nil {
		return nil, err
	}

	hdr := dict{"Content-Type": mpWriter.FormDataContentType()}
	var res *api.DriveUploadFileResult
	if err := d.c.post(contentURL, body, hdr, &res); err != nil {
		return nil, err
	}
	return d.updateContentWs(folderID, res, docID, name, mtime)
}

// getUploadContentWsURL returns the contentWS endpoint URL to add a new file
func (d *DriveService) getUploadContentWsURL(name string, mimeType string, size int64) (string, string, error) {
	token := d.getTokenFromCookie()
	if token == "" {
		return "", "", errors.New("cannot obtain upload token")
	}
	url := d.docRoot + "/ws/com.apple.CloudDocs/upload/web?token=" + token

	data := dict{
		"filename":     name,
		"type":         "FILE",
		"content_type": mimeType,
		"size":         size,
	}
	hdr := dict{"Content-Type": "text/plain"} // sic!

	var (
		res           []api.DriveUploadContentWsResult
		docID, docURL string
	)
	if err := d.c.post(url, data, hdr, &res); err != nil {
		return "", "", err
	}
	if len(res) > 0 {
		docID = res[0].DocID
		docURL = res[0].URL
	}
	if docID == "" || docURL == "" {
		return "", "", errors.New("invalid content url")
	}
	return docID, docURL, nil
}

func (d *DriveService) updateContentWs(folderID string, uploadResult *api.DriveUploadFileResult, docID string, path string, mtime time.Time) (dict, error) {
	fi := &uploadResult.SingleFile
	baseData := dict{
		"signature":           fi.FileChecksum,
		"wrapping_key":        fi.WrappingKey,
		"reference_signature": fi.ReferenceChecksum,
		"size":                fi.Size,
	}
	data := dict{
		"data":              baseData,
		"command":           "add_file",
		"create_short_guid": true,
		"document_id":       docID,
		"path": dict{
			"starting_document_id": folderID,
			"path":                 path,
		},
		"allow_conflict": true,
		"file_flags": dict{
			"is_writable":   true,
			"is_executable": false,
			"is_hidden":     false,
		},
		"mtime": mtime.UnixMilli(),
		"btime": mtime.UnixMilli(),
	}

	// Add the receipt if we have one (absent for empty files)
	if fi.Receipt != "" {
		baseData["receipt"] = fi.Receipt
	}

	url := d.docRoot + "/ws/com.apple.CloudDocs/update/documents"
	hdr := dict{"Content-Type": "text/plain"} // sic!
	var res dict
	if err := d.c.post(url, data, hdr, &res); res != nil {
		return nil, err
	}
	return res, nil
}

// getTokenFromCookie returns the drive service token
func (d *DriveService) getTokenFromCookie() string {
	const icloudComURL = "https://icloud.com"
	u, _ := url.Parse(icloudComURL)
	re := regexp.MustCompile(`\bt=([^:]+)`)
	for _, cookie := range d.c.Client.Jar.Cookies(u) {
		if cookie.Name == "X-APPLE-WEBAUTH-VALIDATE" {
			if match := re.FindStringSubmatch(cookie.Value); match != nil {
				return match[1]
			}
		}
	}
	return ""
}

// Delete an iCloud Drive item
func (n *DriveNode) Delete() error {
	return n.d.moveToTrash(n.i.DriveID, n.i.Etag)
}

// moveToTrash moves items to trash bin
func (d *DriveService) moveToTrash(nodeID, etag string) error {
	nodeData := dict{
		"drivewsid": nodeID,
		"etag":      etag,
		"clientId":  d.c.session.ClientID,
	}
	data := dict{
		"items": []dict{nodeData},
	}
	return d.c.post(d.svcRoot+"/moveItemsToTrash", data, nil, nil)
}

// Mkdir creates new directory
func (n *DriveNode) Mkdir(folder string) error {
	return n.d.createFolders(n.i.DriveID, folder)
}

func (d *DriveService) createFolders(parent, name string) error {
	folder := dict{
		"clientId": d.c.session.ClientID,
		"name":     name,
	}
	data := dict{
		"destinationDrivewsId": parent,
		"folders":              []dict{folder},
	}
	hdr := dict{"Content-Type": "text/plain"}
	return d.c.post(d.svcRoot+"/createFolders", data, hdr, nil)
}

// Rename a node
func (n *DriveNode) Rename(newName string) error {
	return n.d.renameItems(n.i.DriveID, n.i.Etag, newName)
}

func (d *DriveService) renameItems(nodeID, etag, name string) error {
	node := dict{
		"drivewsid": nodeID,
		"etag":      etag, "name": name,
	}
	data := dict{
		"items": []dict{node},
	}

	return d.c.post(d.svcRoot+"/renameItems", data, nil, nil)
}
