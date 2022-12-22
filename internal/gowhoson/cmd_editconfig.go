package gowhoson

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/tai-ga/gowhoson/pkg/whoson"
	"github.com/urfave/cli/v2"
)

func cmdEditConfig(c *cli.Context) error {
	file := filepath.Join(GetClientConfigDir(), ClientConfig)
	e := NewFileEdit(file)
	e.Edit()

	return nil
}

// FileEdit hold information for editablefile.
type FileEdit struct {
	file   string
	Editor string
	Reader io.Reader
	Writer io.Writer
}

// NewFileEdit return new FileEdit struct pointer.
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

// SetFile set file path to variable.
func (f *FileEdit) SetFile(file string) {
	f.file = file
}

// Edit start edit command.
func (f *FileEdit) Edit() error {
	cmd := f.Editor + " " + f.file
	return f.run(cmd, f.Reader, f.Writer)
}

// GetClientConfigDir return config file directory.
func GetClientConfigDir() string {
	dir := os.Getenv("HOME")
	dir = filepath.Join(dir, ".config", "gowhoson")
	return dir
}

// GetClientConfig return client config file and new ClientConfig struct pointer and error.
func GetClientConfig(c *cli.Context) (string, *whoson.ClientConfig, error) {
	var file string
	if c.String("config") == "" {
		dir := GetClientConfigDir()
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", nil, err
		}
		file = filepath.Join(dir, ClientConfig)
	} else {
		file = c.String("config")
	}
	b, err := os.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return "", nil, err
	}
	config := &whoson.ClientConfig{
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
