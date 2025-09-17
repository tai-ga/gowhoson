package gowhoson

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tai-ga/gowhoson/pkg/whoson"
	"github.com/urfave/cli/v3"
)

func cmdDump(ctx context.Context, c *cli.Command) error {
	var err error
	var sc *whoson.ServerCtl
	config := c.Root().Metadata["config"].(*whoson.ServerCtlConfig)

	config.EditConfig = c.Bool("editconfig")

	if config.EditConfig {
		file := filepath.Join(GetServerCtlConfigDir(), ServerCtlConfig)
		e := NewFileEdit(file)
		err = e.Edit()
		goto DONE
	}

	if c.String("server") != "" {
		config.Server = c.String("server")
	}

	config.JSON = c.Bool("json")

	sc = whoson.NewServerCtl(config.Server)
	sc.SetWriter(c.Root().Writer)
	err = sc.Dump()
	if err != nil {
		return err
	}

	if config.JSON {
		err = sc.WriteJSON()
	} else {
		err = sc.WriteTable()
	}

DONE:
	if err != nil {
		return err
	}
	return nil
}

// GetServerCtlConfigDir return config file directory.
func GetServerCtlConfigDir() string {
	dir := os.Getenv("HOME")
	dir = filepath.Join(dir, ".config", "gowhoson")
	return dir
}

// GetServerCtlConfig return server config file and new ServerCtlConfig struct pointer and error.
func GetServerCtlConfig(c *cli.Command) (string, *whoson.ServerCtlConfig, error) {
	var file string
	if c.String("config") == "" {
		dir := GetServerCtlConfigDir()
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", nil, err
		}
		file = filepath.Join(dir, ServerCtlConfig)
	} else {
		file = c.String("config")
	}
	b, err := os.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return "", nil, err
	}
	config := &whoson.ServerCtlConfig{
		Server: "127.0.0.1:9877",
		JSON:   false,
	}
	if err == nil {
		err = json.Unmarshal(b, &config)
		if err != nil {
			return "", nil, fmt.Errorf("could not unmarshal %v: %v", file, err)
		}
	}
	return file, config, nil
}
