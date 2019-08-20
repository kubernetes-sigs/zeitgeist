// verify that dependencies are up-to-date across different files
package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pluies/zeitgeist/upstreams"

	"github.com/blang/semver"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v3"
)

type Dependencies struct {
	Dependencies []*Dependency `yaml:"dependencies"`
}

type Dependency struct {
	Name     string     `yaml:"name"`
	Version  string     `yaml:"version"`
	Upstream *Upstream  `yaml:"upstream"`
	Semver   bool       `yaml:"semver"`
	RefPaths []*RefPath `yaml:"refPaths"`
}

type RefPath struct {
	Path  string `yaml:"path"`
	Match string `yaml:"match"`
}

type Upstream struct {
	Flavour     upstreams.UpstreamFlavour `yaml:"flavour"`
	URL         string                    `yaml:"url"`
	Constraints string                    `yaml:"constraints"`
}

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
			Name:    "local",
			Aliases: []string{},
			Usage:   "Check all local files against declared dependency version",
			Action: func(c *cli.Context) error {
				initLogging(verbose)
				localCheck(config)
				return nil
			},
		},
		{
			Name:    "validate",
			Aliases: []string{},
			Usage:   "Check dependencies locally and against upstream versions",
			Action: func(c *cli.Context) error {
				initLogging(verbose)
				localCheck(config)
				remoteCheck(config, githubAccessToken)
				return nil
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		log.Info("Hello friend!")
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func dependenciesFromFile(dependencyFilePath string) *Dependencies {
	depFile, err := ioutil.ReadFile(dependencyFilePath)
	if err != nil {
		log.Fatal(err)
	}

	dependencies := &Dependencies{}
	err = yaml.Unmarshal(depFile, dependencies)
	if err != nil {
		log.Fatal(err)
	}
	return dependencies
}

func localCheck(dependencyFilePath string) {
	base := filepath.Dir(dependencyFilePath)
	externalDeps := dependenciesFromFile(dependencyFilePath)
	var nonMatchingPaths []string
	for _, dep := range externalDeps.Dependencies {
		log.Debugf("Examining dependency: %v", dep.Name)
		for _, refPath := range dep.RefPaths {
			filePath := filepath.Join(base, refPath.Path)
			file, err := os.Open(filePath)
			if err != nil {
				log.Fatalf("Error opening %v: %v", filePath, err)
			}
			log.Debugf("Examining file: %v", filePath)
			match := refPath.Match
			matcher := regexp.MustCompile(match)
			scanner := bufio.NewScanner(file)

			var found bool
			var lineNumber int
			for scanner.Scan() {
				lineNumber += 1
				line := scanner.Text()
				if matcher.MatchString(line) {
					if strings.Contains(line, dep.Version) {
						log.Debugf("Line %v matches expected regexp '%v' and version '%v':\n%v", lineNumber, match, dep.Version, line)
						found = true
						break
					} else {
						log.Warnf("Line %v matches expected regexp '%v', but not version '%v':\n%v", lineNumber, match, dep.Version, line)
					}
				}
			}
			if !found {
				log.Debugf("Finished reading file %v, no match found.", filePath)
				nonMatchingPaths = append(nonMatchingPaths, refPath.Path)
			}
		}

		if len(nonMatchingPaths) > 0 {
			log.Fatalf("%v indicates that %v should be at version %v, but the following files didn't match:\n\n"+
				"%v\n", dependencyFilePath, dep.Name, dep.Version, strings.Join(nonMatchingPaths, "\n"))
		}
	}
}

func remoteCheck(dependencyFilePath string, githubAccessToken string) {
	externalDeps := dependenciesFromFile(dependencyFilePath)
	for _, dep := range externalDeps.Dependencies {
		if dep.Upstream == nil {
			continue
		}

		log.Debugf("Examining dependency: %v", dep.Name)

		var latestVersion string = dep.Version
		var currentVersion string = dep.Version

		switch dep.Upstream.Flavour {
		case upstreams.GitHub:
			gh := upstreams.Github{
				AccessToken: githubAccessToken,
				URL:         dep.Upstream.URL,
				Constraints: dep.Upstream.Constraints,
			}
			latestVersion = gh.LatestVersion()
		default:
			log.Fatalf("Unknown upstream type '%v' for dependency %v", dep.Upstream.Flavour, dep.Name)
		}

		if Version(latestVersion).MoreRecentThan(Version(currentVersion)) {
			log.Infof("Update available for dependency %v: %v (current: %v)\n", dep.Name, latestVersion, currentVersion)
		} else {
			log.Infof("No update available for dependency %v: %v (latest: %v)\n", dep.Name, currentVersion, latestVersion)
		}
	}
}

type Version string

func (a Version) MoreRecentThan(b Version) bool {
	// Try and parse as Semver first
	semverComparison := true
	aSemver, err := semver.Parse(string(a))
	if err != nil {
		semverComparison = false
	}
	bSemver, err := semver.Parse(string(b))
	if err != nil {
		semverComparison = false
	}
	if semverComparison {
		return aSemver.GT(bSemver)
	} else {
		// Failed semver: fallback to standard string comparison (lexicographic)
		return strings.Compare(string(a), string(b)) > 0
	}
}
