package main

import (
	"os"

	"github.com/ivandeex/go-icloud/icloud"
	icloudapi "github.com/ivandeex/go-icloud/icloud/api"
	"github.com/pkg/errors"
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

	api, err := icloud.NewClient(username, password, "", "")
	if err == nil {
		err = api.Authenticate(false, "")
	}
	if err != nil {
		return err
	}

	if api.Requires2SA() {
		log.Warn("Two-step authentication required.")
		var devices []icloudapi.Device
		if devices, err = api.TrustedDevices(); err != nil {
			return err
		}
		log.Warnf("Your trusted devices are: %#v", devices)
		dev := &devices[0]
		log.Warnf("Sending verification code to the first device...")
		if err = api.SendVerificationCode(dev); err != nil {
			return err
		}
		code := icloud.ReadLine("Please enter validation code: ")
		if err = api.ValidateVerificationCode(dev, code); err != nil {
			return errors.Wrap(err, "failed to verify verification code")
		}
	}

	if api.Requires2FA() {
		log.Warnf("Two-factor authentication required.")
		code := icloud.ReadLine("Enter the code you received of one of your approved devices: ")
		if err = api.Validate2FACode(code); err != nil {
			return errors.Wrap(err, "failed to verify security code")
		}
		if !api.IsTrustedSession() {
			log.Infof("Session is not trusted. Requesting trust...")
			if err = api.TrustSession(); err != nil {
				log.Errorf("Failed to request trust. You will likely be prompted for the code again in the coming weeks")
				return err
			}
		}
	}

	log.Infof("Successfully authenticated")
	return nil
}
