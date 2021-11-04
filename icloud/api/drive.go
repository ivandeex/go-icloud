package api

import (
	"time"
)

// DriveItem ...
type DriveItem struct {
	// Common attributes of folder or file
	Name        string    `json:"name"`
	Ext         string    `json:"extension"`
	Type        string    `json:"type"`
	Size        *int64    `json:"size"`
	Status      string    `json:"status"`
	Zone        string    `json:"zone"`
	Created     time.Time `json:"dateCreated"`
	Changed     time.Time `json:"dateChanged"`
	Modified    time.Time `json:"dateModified"`
	LastOpened  time.Time `json:"lastOpenTime"`
	Quota       int       `json:"assetQuota"`
	DocID       string    `json:"docwsid"`
	DriveID     string    `json:"drivewsid"`
	ParentID    string    `json:"parentId"`
	Chained     bool      `json:"isChainedToParent"`
	Etag        string    `json:"etag"`
	AliasCount  int       `json:"shareAliasCount"`
	ShareCount  int       `json:"shareCount"`
	DirectCount int       `json:"directChildrenCount"`
	FileCount   int       `json:"fileCount"`
	ItemCount   int       `json:"numberOfItems"`
	// Children (folder only)
	Items []*DriveItem `json:"items"`
}

type DriveDocResult struct {
	DataToken struct {
		URL string `json:"url"`
	} `json:"data_token"`
}

type DriveUploadContentWsResult struct {
	DocID   string `json:"document_id"`
	Owner   string `json:"owner"`
	OwnerID string `json:"owner_id"`
	URL     string `json:"url"`
}

type DriveUploadFileResult struct {
	SingleFile struct {
		FileChecksum      string `json:"fileChecksum"`
		ReferenceChecksum string `json:"referenceChecksum"`
		Receipt           string `json:"receipt"`
		Size              int64  `json:"size"`
		WrappingKey       string `json:"wrappingKey"`
	} `json:"singleFile"`
}
