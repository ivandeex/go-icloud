package main

import (
	"os"

	"github.com/ivandeex/go-icloud/icloud"
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
		PadLevelText:    true,
		FullTimestamp:   true,
		TimestampFormat: "15:04:05.999",
	})
	icloud.Debug = level >= log.DebugLevel

	if username == "" || password == "" {
		return errors.New("username or password was not supplied")
	}

	api, err := icloud.New(username, password)
	if err == nil {
		err = api.Authenticate(false, "")
	}
	if err != nil {
		return err
	}

	return nil
}
