package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/ivandeex/go-icloud/icloud"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	username string
	password string
	verbose  int
)

func init() {
	// Globally disable alphabetical sorting of all commands in help output.
	cobra.EnableCommandSorting = false

	// Disable alphabetical sorting of flags in help output.
	flags := rootCommand.Flags()
	flags.SortFlags = false

	flags.StringVarP(&username, "username", "u", username, "Apple ID to use")
	flags.StringVarP(&password, "password", "p", password, "Apple ID password to use")
	flags.CountVarP(&verbose, "verbose", "v", "Log more stuff")
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCommand = &cobra.Command{
	Use:          "icloud",
	Short:        "Apple iCloud CLI",
	RunE:         rootMain,
	SilenceUsage: true,
}

func rootMain(command *cobra.Command, _ []string) error {
	if verbose < 0 {
		verbose = 0
	}
	var level log.Level
	switch verbose {
	case 0:
		level = log.ErrorLevel
	case 1:
		level = log.InfoLevel
	case 2:
		level = log.DebugLevel
	default:
		level = log.TraceLevel
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		DisableQuote:    true,
		FullTimestamp:   true,
		TimestampFormat: "15:04:05.999",
	})
	icloud.Debug = level >= log.DebugLevel

	if username == "" || password == "" {
		return errors.New("username or password was not supplied")
	}

	cli, err := icloud.NewClient(username, password, "", "")
	if err == nil {
		err = cli.Authenticate(false, "")
	}
	if err != nil {
		return err
	}

	if cli.Requires2SA() {
		log.Warn("Two-step authentication required.")
		devices, err := cli.TrustedDevices()
		if err != nil {
			return err
		}
		log.Warnf("Your trusted devices are: %#v", devices)
		dev := &devices[0]
		log.Warnf("Sending verification code to the first device...")
		if err = cli.SendVerificationCode(dev); err != nil {
			return err
		}
		code := icloud.ReadLine("Please enter validation code: ")
		if err = cli.ValidateVerificationCode(dev, code); err != nil {
			return fmt.Errorf("failed to verify verification code: %w", err)
		}
	}

	if cli.Requires2FA() {
		log.Warnf("Two-factor authentication required.")
		code := icloud.ReadLine("Enter the code you received of one of your approved devices: ")
		if err = cli.Validate2FACode(code); err != nil {
			return fmt.Errorf("failed to verify security code: %w", err)
		}
		if !cli.IsTrustedSession() {
			log.Infof("Session is not trusted. Requesting trust...")
			if err = cli.TrustSession(); err != nil {
				log.Errorf("Failed to request trust. You will likely be prompted for the code again in the coming weeks")
				return err
			}
		}
	}

	log.Infof("Successfully authenticated")

	drive, err := icloud.NewDrive(cli)
	if err != nil {
		return fmt.Errorf("cannot connect to drive service: %w", err)
	}
	root, err := drive.Root()
	if err != nil {
		return fmt.Errorf("cannot obtain iDrive root: %w", err)
	}
	dir, _ := root.Dir()
	log.Infof("root name %q type %q dir %q", root.Name(), root.Type(), dir)
	if len(dir) == 0 {
		return errors.New("root folder is empty")
	}

	subdir, err := root.Get(dir[0])
	if err != nil {
		return fmt.Errorf("%s: cannot read subdir: %w", dir[0], err)
	}
	dir, _ = subdir.Dir()
	log.Infof("subdir name %q dir %q", subdir.Name(), dir)
	if len(dir) == 0 {
		return fmt.Errorf("%s: folder is empty", subdir.Name())
	}

	name, path := dir[0], "test.log"
	file, err := subdir.Get(name)
	if err != nil {
		return fmt.Errorf("file %q not found in folder %q: %w", name, subdir.Name(), err)
	}
	err = file.Download(path)
	if err != nil {
		return fmt.Errorf("%s: cannot download: %w", name, err)
	}
	log.Infof("saved %q into %q", name, path)

	reLog := regexp.MustCompile(`^test.*\.log$`)
	for _, name := range dir {
		if reLog.MatchString(name) {
			file, err = subdir.Get(name)
			if err == nil {
				err = file.Delete()
			}
			log.Infof("removing %q returns %v", name, err)
		}
	}
	err = subdir.Upload(path)
	log.Infof("uploading %q returns %v", path, err)

	return err
}
