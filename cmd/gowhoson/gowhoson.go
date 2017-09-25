package gowhoson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
)

// Versions hold information for program and git and go versions.
type Versions struct {
	Version   string
	Gitcommit string
	Goversion string
}

// NewVersions return new Versions struct pointer.
func NewVersions(v, git, gov string) *Versions {
	return &Versions{
		Version:   v,
		Gitcommit: git,
		Goversion: gov,
	}
}

var (
	// AppVersions is a application version.
	AppVersions *Versions
)

func makeApp() *cli.App {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version %s, build %s, Go:%s\n", c.App.Name, c.App.Version, AppVersions.Gitcommit, AppVersions.Goversion)
	}

	app := cli.NewApp()
	app.Name = "gowhoson"
	app.Usage = "gowhoson is whoson server & client"
	app.Version = AppVersions.Version
	app.ErrWriter = os.Stderr
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config",
			Usage:  "config file path",
			Value:  "",
			EnvVar: "GOWHOSON_CONFIG",
		},
	}

	clientFlags := []cli.Flag{
		cli.StringFlag{
			Name:   "mode",
			Usage:  "e.g. [tcp|udp]",
			EnvVar: "GOWHOSON_CLIENT_MODE",
		},
		cli.StringFlag{
			Name:   "server",
			Usage:  "e.g. [ServerIP:Port]",
			EnvVar: "GOWHOSON_CLIENT_SERVER",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "server",
			Usage: "gowhoson server mode",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "tcp",
					Usage:  "e.g. [ServerIP:Port|nostart]",
					EnvVar: "GOWHOSON_SERVER_TCP",
				},
				cli.StringFlag{
					Name:   "udp",
					Usage:  "e.g. [ServerIP:Port|nostart]",
					EnvVar: "GOWHOSON_SERVER_UDP",
				},
				cli.StringFlag{
					Name:   "log",
					Usage:  "e.g. [stdout|stderr|discard] or \"/var/log/filename.log\"",
					EnvVar: "GOWHOSON_SERVER_LOG",
				},
				cli.StringFlag{
					Name:   "loglevel",
					Usage:  "e.g. [debug|info|warn|error|dpanic|panic|fatal]",
					EnvVar: "GOWHOSON_SERVER_LOGLEVEL",
				},
				cli.IntFlag{
					Name:   "serverid",
					Usage:  "e.g. [1000]",
					EnvVar: "GOWHOSON_SERVER_SERVERID",
				},
				cli.BoolFlag{
					Name:   "expvar",
					Usage:  "e.g. (default: false)",
					EnvVar: "GOWHOSON_SERVER_EXPVAR",
				},
				cli.StringFlag{
					Name:   "controlport",
					Usage:  "e.g. [ServerIP:Port]",
					EnvVar: "GOWHOSON_SERVER_CONTROLPORT",
				},
				cli.StringFlag{
					Name:   "syncremote",
					Usage:  "e.g. [ServerIP:Port,ServerIP:Port...]",
					EnvVar: "GOWHOSON_SERVER_SYNCREMOTE",
				},
				cli.StringFlag{
					Name:   "savefile",
					Usage:  "e.g. [/var/lib/gowhoson.json]",
					EnvVar: "GOWHOSON_SERVER_SAVEFILE",
				},
			},
			Action: cmdServer,
		},
		{
			Name:  "client",
			Usage: "gowhoson client mode",
			Subcommands: []cli.Command{
				{
					Name:   "login",
					Usage:  "whoson command \"LOGIN\"",
					Flags:  clientFlags,
					Action: cmdLogin,
				},
				{
					Name:   "query",
					Usage:  "whoson command \"QUERY\"",
					Flags:  clientFlags,
					Action: cmdQuery,
				},
				{
					Name:   "logout",
					Usage:  "whoson command \"LOGOUT\"",
					Flags:  clientFlags,
					Action: cmdLogout,
				},
				{
					Name:   "editconfig",
					Usage:  "edit client configration file",
					Action: cmdEditConfig,
				},
			},
		},
		{
			Name:  "dump",
			Usage: "gowhoson server control dump mode",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "server",
					Usage:  "e.g. [ServerIP:Port]",
					EnvVar: "GOWHOSON_SERVERCTL_DUMP_SERVER",
				},
				cli.BoolFlag{
					Name:   "json",
					Usage:  "e.g. (default: false)",
					EnvVar: "GOWHOSON_SERVERCTL_DUMP_JSON",
				},
			},
			Action: cmdDump,
		},
	}
	app.Setup()
	return app
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Run cmd/gowhoson package entry point.
func Run() int {
	app := makeApp()

	app.Before = func(c *cli.Context) error {
		if len(c.Args()) > 0 && c.Args()[0] == "client" {
			err := runClient(c, app)
			if err != nil {
				return err
			}
		} else if len(c.Args()) > 0 && c.Args()[0] == "dump" {
			err := runDump(c, app)
			if err != nil {
				return err
			}
		} else if len(c.Args()) > 0 && c.Args()[0] == "server" {
			err := runServer(c, app)
			if err != nil {
				return err
			}
		}
		return nil
	}

	app.Run(os.Args)
	return 0
}

func runClient(c *cli.Context, app *cli.App) error {
	file, config, err := GetClientConfig(c)
	if err != nil {
		return err
	}
	app.Metadata = map[string]interface{}{
		"config": config,
	}

	if !fileExists(file) {
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
		err = ioutil.WriteFile(file, b, 0600)
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
	}
	return nil
}

func runDump(c *cli.Context, app *cli.App) error {
	file, config, err := GetServerCtlConfig(c)
	if err != nil {
		return err
	}
	app.Metadata = map[string]interface{}{
		"config": config,
	}

	if !fileExists(file) {
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
		err = ioutil.WriteFile(file, b, 0600)
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
	}
	return nil
}

func runServer(c *cli.Context, app *cli.App) error {
	file, config, err := GetServerConfig(c)
	if err != nil {
		return err
	}
	app.Metadata = map[string]interface{}{
		"config": config,
	}

	if !fileExists(file) {
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
		err = ioutil.WriteFile(file, b, 0644)
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
	}
	return nil
}
