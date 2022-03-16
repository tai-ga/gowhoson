package gowhoson

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/tai-ga/gowhoson/pkg/whoson"
	"github.com/urfave/cli/v2"
)

// GetServerConfig return server config file and new ServerConfig struct pointer and error.
func GetServerConfig(c *cli.Context) (string, *whoson.ServerConfig, error) {
	var file string
	if c.String("config") == "" {
		file = filepath.Join("/etc", ServerConfig)
	} else {
		file = c.String("config")
	}

	b, err := ioutil.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return "", nil, err
	}
	config := &whoson.ServerConfig{
		TCP:        "127.0.0.1:9876",
		UDP:        "127.0.0.1:9876",
		Log:        "stdout",
		Loglevel:   "error",
		ServerID:   1000,
		Expvar:     false,
		SyncRemote: "",
		SaveFile:   "",
	}
	if err == nil {
		err = json.Unmarshal(b, &config)
		if err != nil {
			return "", nil, fmt.Errorf("could not unmarshal %v: %v", file, err)
		}
	}
	return file, config, nil
}

func optOverwite(c *cli.Context, config *whoson.ClientConfig) {
	if c.String("mode") != "" {
		config.Mode = c.String("mode")
	}
	if c.String("server") != "" {
		config.Server = c.String("server")
	}
}

func displayError(w io.Writer, e error) {
	//color.Set(color.FgYellow)
	fmt.Fprintln(w, e.Error())
	//color.Set(color.Reset)
}

func display(w io.Writer, s string) {
	fmt.Fprintln(w, s)
}
