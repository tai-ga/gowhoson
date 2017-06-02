package gowhoson

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli"
)

type FileEdit struct {
	file   string
	Editor string
	Reader io.Reader
	Writer io.Writer
}

func NewFileEdit(file string) *FileEdit {
	e := &FileEdit{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
	e.setEditor()
	e.SetFile(file)
	return e
}

func (f *FileEdit) run(command string, r io.Reader, w io.Writer) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = f.Writer
	cmd.Stdin = f.Reader
	return cmd.Run()
}

func (f *FileEdit) setEditor() {
	f.Editor = os.Getenv("EDITOR")
	if f.Editor == "" && runtime.GOOS != "windows" {
		f.Editor = "vi"
	}
}

func (f *FileEdit) SetFile(file string) {
	f.file = file
}

func (f *FileEdit) Edit() error {
	cmd := f.Editor + " " + f.file
	return f.run(cmd, f.Reader, f.Writer)
}

func GetClientConfigDir() string {
	dir := os.Getenv("HOME")
	dir = filepath.Join(dir, ".config", "gowhoson")
	return dir
}

func GetClientConfig(c *cli.Context) (string, *ClientConfig, error) {
	var file string
	if c.String("config") == "" {
		dir := GetClientConfigDir()
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", nil, err
		}
		file = filepath.Join(dir, CLIENT_CONFIG)
	} else {
		file = c.String("config")
	}
	b, err := ioutil.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return "", nil, err
	}
	config := &ClientConfig{
		Mode:   "udp",
		Server: "127.0.0.1:9876",
	}
	if err == nil {
		err = json.Unmarshal(b, &config)
		if err != nil {
			return "", nil, fmt.Errorf("could not unmarshal %v: %v", file, err)
		}
	}
	return file, config, nil
}

func GetServerConfig(c *cli.Context) (string, *ServerConfig, error) {
	var file string
	if c.String("config") == "" {
		file = filepath.Join("/etc", SERVER_CONFIG)
	} else {
		file = c.String("config")
	}

	b, err := ioutil.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return "", nil, err
	}
	config := &ServerConfig{
		TCP: "127.0.0.1:9876",
		UDP: "127.0.0.1:9876",
	}
	if err == nil {
		err = json.Unmarshal(b, &config)
		if err != nil {
			return "", nil, fmt.Errorf("could not unmarshal %v: %v", file, err)
		}
	}
	return file, config, nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func optOverwite(c *cli.Context, config *ClientConfig) {
	if c.String("mode") != "" {
		config.Mode = c.String("mode")
	}
	if c.String("server") != "" {
		config.Server = c.String("server")
	}
}

func optOverwiteServer(c *cli.Context, config *ServerConfig) {
	if c.String("tcp") != "" {
		config.TCP = c.String("tcp")
	}
	if c.String("udp") != "" {
		config.UDP = c.String("udp")
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
