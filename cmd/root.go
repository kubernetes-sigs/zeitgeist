/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/* Zeitgeist is a is a language-agnostic dependency checker */

package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"sigs.k8s.io/zeitgeist/dependencies"
)

// Variables set by GoReleaser on release
var (
	version = "dev"
	commit  = "none"    // nolint: deadcode,gochecknoglobals,unused,varcheck
	date    = "unknown" // nolint: deadcode,gochecknoglobals,unused,varcheck
)

// Initialise logging level based on LOG_LEVEL env var, or the --verbose flag.
// Defaults to info
func initLogging(verbose, json bool) {
	logLevelStr, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		if verbose {
			logLevelStr = "debug"
		} else {
			logLevelStr = "info"
		}
	}

	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		log.Fatalf("Invalid LOG_LEVEL: %v", logLevelStr)
	}

	log.SetLevel(logLevel)

	if json {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(
			&log.TextFormatter{
				FullTimestamp:          true,
				DisableLevelTruncation: true,
			},
		)
	}
}

func main() {
	var (
		verbose bool
		json    bool
		config  string
	)

	app := cli.NewApp()
	app.Name = "zeitgeist"
	app.Usage = "Manage your external dependencies"
	app.Version = fmt.Sprintf("%v", version)
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "Set log level to DEBUG",
			Destination: &verbose,
		},
		cli.BoolFlag{
			Name:        "json-output",
			Usage:       "JSON logging output",
			Destination: &json,
		},
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Load configuration from `FILE`",
			Value:       "dependencies.yaml",
			Destination: &config,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "validate",
			Aliases: []string{},
			Usage:   "Check dependencies locally and against upstream versions",
			Action: func(c *cli.Context) error {
				initLogging(verbose, json)
				err := dependencies.LocalCheck(config)
				if err != nil {
					return err
				}
				updates, err := dependencies.RemoteCheck(config)
				if err != nil {
					return err
				}
				for _, update := range updates {
					fmt.Printf(update + "\n")
				}
				return nil
			},
		},
		{
			Name:    "local",
			Aliases: []string{},
			Usage:   "Only check dependency consistency locally",
			Action: func(c *cli.Context) error {
				initLogging(verbose, json)
				return dependencies.LocalCheck(config)
			},
		},
	}

	// Default action when no action is passed: display the help
	// nolint: gocritic
	app.Action = func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
