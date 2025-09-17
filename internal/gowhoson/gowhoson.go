package gowhoson

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v3"
)

// Versions hold information for program and git and go versions.
type Versions struct {
	Version   string
	Gitcommit string
}

// NewVersions return new Versions struct pointer.
func NewVersions(v, git string) *Versions {
	return &Versions{
		Version:   v,
		Gitcommit: git,
	}
}

var (
	// AppVersions is a application version.
	AppVersions *Versions
)

func makeApp() *cli.Command {
	cli.VersionPrinter = func(c *cli.Command) {
		fmt.Printf("%s version:%s, build:%s, Go:%s\n", c.Name, c.Version, AppVersions.Gitcommit, runtime.Version())
	}

	app := &cli.Command{
		Name:      "gowhoson",
		Usage:     "gowhoson is whoson server & client",
		Version:   AppVersions.Version,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Usage:   "config file path",
			Value:   "",
			Sources: cli.EnvVars("GOWHOSON_CONFIG"),
		},
	}

	clientFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "mode",
			Usage:   "e.g. [tcp|udp]",
			Sources: cli.EnvVars("GOWHOSON_CLIENT_MODE"),
		},
		&cli.StringFlag{
			Name:    "server",
			Usage:   "e.g. [ServerIP:Port]",
			Sources: cli.EnvVars("GOWHOSON_CLIENT_SERVER"),
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:  "server",
			Usage: "gowhoson server mode",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "tcp",
					Usage:   "e.g. [ServerIP:Port|nostart]",
					Sources: cli.EnvVars("GOWHOSON_SERVER_TCP"),
				},
				&cli.StringFlag{
					Name:    "udp",
					Usage:   "e.g. [ServerIP:Port|nostart]",
					Sources: cli.EnvVars("GOWHOSON_SERVER_UDP"),
				},
				&cli.StringFlag{
					Name:    "log",
					Usage:   "e.g. [stdout|stderr|discard] or \"/var/log/filename.log\"",
					Sources: cli.EnvVars("GOWHOSON_SERVER_LOG"),
				},
				&cli.StringFlag{
					Name:    "loglevel",
					Usage:   "e.g. [debug|info|warn|error|dpanic|panic|fatal]",
					Sources: cli.EnvVars("GOWHOSON_SERVER_LOGLEVEL"),
				},
				&cli.IntFlag{
					Name:    "serverid",
					Usage:   "e.g. [1000]",
					Sources: cli.EnvVars("GOWHOSON_SERVER_SERVERID"),
				},
				&cli.BoolFlag{
					Name:    "expvar",
					Usage:   "e.g. (default: false)",
					Sources: cli.EnvVars("GOWHOSON_SERVER_EXPVAR"),
				},
				&cli.StringFlag{
					Name:    "controlport",
					Usage:   "e.g. [ServerIP:Port]",
					Sources: cli.EnvVars("GOWHOSON_SERVER_CONTROLPORT"),
				},
				&cli.StringFlag{
					Name:    "syncremote",
					Usage:   "e.g. [ServerIP:Port,ServerIP:Port...]",
					Sources: cli.EnvVars("GOWHOSON_SERVER_SYNCREMOTE"),
				},
				&cli.StringFlag{
					Name:    "savefile",
					Usage:   "e.g. [/var/lib/gowhoson.json]",
					Sources: cli.EnvVars("GOWHOSON_SERVER_SAVEFILE"),
				},
			},
			Action: cmdServer,
		},
		{
			Name:  "client",
			Usage: "gowhoson client mode",
			Commands: []*cli.Command{
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
				&cli.StringFlag{
					Name:    "server",
					Usage:   "e.g. [ServerIP:Port]",
					Sources: cli.EnvVars("GOWHOSON_SERVERCTL_DUMP_SERVER"),
				},
				&cli.BoolFlag{
					Name:    "json",
					Usage:   "e.g. (default: false)",
					Sources: cli.EnvVars("GOWHOSON_SERVERCTL_DUMP_JSON"),
				},
				&cli.BoolFlag{
					Name:    "editconfig",
					Usage:   "e.g. (default: false)",
					Sources: cli.EnvVars("GOWHOSON_SERVERCTL_DUMP_EDITCONFIG"),
				},
			},
			Action: cmdDump,
		},
	}
	return app
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Run cmd/gowhoson package entry point.
func Run() int {
	app := makeApp()

	app.Before = func(ctx context.Context, c *cli.Command) (context.Context, error) {
		if c.Args().Len() > 0 && c.Args().Slice()[0] == "client" {
			err := runClient(ctx, c, app)
			if err != nil {
				return ctx, err
			}
		} else if c.Args().Len() > 0 && c.Args().Slice()[0] == "dump" {
			err := runDump(ctx, c, app)
			if err != nil {
				return ctx, err
			}
		} else if c.Args().Len() > 0 && c.Args().Slice()[0] == "server" {
			err := runServer(ctx, c, app)
			if err != nil {
				return ctx, err
			}
		}
		return ctx, nil
	}

	app.Run(context.Background(), os.Args)
	return 0
}

func runClient(_ context.Context, c *cli.Command, app *cli.Command) error {
	file, config, err := GetClientConfig(c)
	if err != nil {
		return err
	}
	app.Metadata = map[string]any{
		"config": config,
	}

	if !fileExists(file) {
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
		err = os.WriteFile(file, b, 0600)
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
	}
	return nil
}

func runDump(_ context.Context, c *cli.Command, app *cli.Command) error {
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
		err = os.WriteFile(file, b, 0600)
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
	}
	return nil
}

func runServer(_ context.Context, c *cli.Command, app *cli.Command) error {
	file, config, err := GetServerConfig(c)
	if err != nil {
		return err
	}
	app.Metadata = map[string]any{
		"config": config,
	}

	if !fileExists(file) {
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
		err = os.WriteFile(file, b, 0644)
		if err != nil {
			return fmt.Errorf("failed to store file: %v", err)
		}
	}
	return nil
}
