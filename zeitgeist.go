// verify that dependencies are up-to-date across different files
package main

import (
	"os"

	"github.com/pluies/zeitgeist/dependencies"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Initialise logging level based on LOG_LEVEL env var, or the --verbose flag.
// Defaults to info
func initLogging(verbose bool) {
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
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	})
}

func main() {
	var verbose bool
	var config string
	var githubAccessToken string

	app := cli.NewApp()
	app.Name = "zeitgeist"
	app.Usage = "Manage your external dependencies"
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "Set log level to DEBUG",
			Destination: &verbose,
		},
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Load configuration from `FILE`",
			Value:       "dependencies.yaml",
			Destination: &config,
		},
		cli.StringFlag{
			Name:        "github-access-token",
			Usage:       "Access token to use when querying the Github API",
			EnvVar:      "GITHUB_ACCESS_TOKEN",
			Destination: &githubAccessToken,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "validate",
			Aliases: []string{},
			Usage:   "Check dependencies locally and against upstream versions",
			Action: func(c *cli.Context) error {
				initLogging(verbose)
				err := dependencies.LocalCheck(config)
				if err != nil {
					return err
				}
				return dependencies.RemoteCheck(config, githubAccessToken)
			},
		},
		{
			Name:    "local",
			Aliases: []string{},
			Usage:   "Only check dependency consistency locally",
			Action: func(c *cli.Context) error {
				initLogging(verbose)
				return dependencies.LocalCheck(config)
			},
		},
	}

	// Default action when no action is passed: display the help
	app.Action = func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
