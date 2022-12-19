package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mrtc0/appman/application"
	"github.com/mrtc0/appman/application/config"
	"github.com/urfave/cli/v2"
)

const refreshInterval = 1 * time.Second

func main() {
	logger := log.Default()
	app := &cli.App{
		Name:  "appman",
		Usage: "appman - Application Manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "appman.yaml",
				Usage:   "Load configuration from `FILE`",
			},
		},
		Action: func(cCtx *cli.Context) error {
			configPath := cCtx.String("config")
			if configPath == "" {
				return fmt.Errorf("configuration file is not specified")
			}

			applicationConfig, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("invalid configuration: %s", err)
			}
			manager := application.NewTuiApplicationManager(*applicationConfig)
			go manager.Refresh(refreshInterval)
			if err := manager.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}
