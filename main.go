package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tai-ga/gowhoson/cmd/gowhoson"
	"github.com/urfave/cli"
)

var (
	gVersion   string
	gGitcommit string
	gGoversion string
)

func makeApp() *cli.App {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version %s, build %s, Go:%s\n", c.App.Name, c.App.Version, gGitcommit, gGoversion)
	}

	app := cli.NewApp()
	app.Name = "gowhoson"
	app.Usage = "gowhoson is whoson server & client"
	app.Version = gVersion
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
			},
			Action: gowhoson.CmdServer,
		},
		{
			Name:  "client",
			Usage: "gowhoson client mode",
			Subcommands: []cli.Command{
				{
					Name:   "login",
					Usage:  "whoson command \"LOGIN\"",
					Flags:  clientFlags,
					Action: gowhoson.CmdLogin,
				},
				{
					Name:   "query",
					Usage:  "whoson command \"QUERY\"",
					Flags:  clientFlags,
					Action: gowhoson.CmdQuery,
				},
				{
					Name:   "logout",
					Usage:  "whoson command \"LOGOUT\"",
					Flags:  clientFlags,
					Action: gowhoson.CmdLogout,
				},
				{
					Name:   "editconfig",
					Usage:  "edit client configration file",
					Action: gowhoson.CmdEditConfig,
				},
			},
		},
	}
	app.Setup()
	return app
}

func run() int {
	app := makeApp()

	app.Before = func(c *cli.Context) error {
		if len(c.Args()) > 0 && c.Args()[0] == "client" {
			file, config, err := gowhoson.GetClientConfig(c)
			if err != nil {
				return err
			}
			app.Metadata = map[string]interface{}{
				"config": config,
			}

			if !gowhoson.FileExists(file) {
				b, err := json.MarshalIndent(config, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to store file: %v", err)
				}
				err = ioutil.WriteFile(file, b, 0600)
				if err != nil {
					return fmt.Errorf("failed to store file: %v", err)
				}
			}
		} else {
			file, config, err := gowhoson.GetServerConfig(c)
			if err != nil {
				return err
			}
			app.Metadata = map[string]interface{}{
				"config": config,
			}

			if !gowhoson.FileExists(file) {
				b, err := json.MarshalIndent(config, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to store file: %v", err)
				}
				err = ioutil.WriteFile(file, b, 0644)
				if err != nil {
					return fmt.Errorf("failed to store file: %v", err)
				}
			}
		}
		return nil
	}

	app.Run(os.Args)
	return 0
}

func main() {
	os.Exit(run())
}
