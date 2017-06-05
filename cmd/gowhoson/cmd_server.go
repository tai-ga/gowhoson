package gowhoson

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/tai-ga/gowhoson/whoson"
	"github.com/urfave/cli"
)

func signalHandler(ch <-chan os.Signal, wg *sync.WaitGroup, c *cli.Context, f func()) {
	defer wg.Done()

	<-ch
	f()
	time.AfterFunc(time.Second*8, func() {
		displayError(c.App.ErrWriter, errors.New("Clean shutdown took too long, forcing exit"))
		os.Exit(0)
	})
}

func splitHostPort(hostPort string) (host string, port int, err error) {
	host, p, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", 0, err
	}
	port, err = strconv.Atoi(p)
	if err != nil {
		return "", 0, err
	}
	return
}

func optOverwiteServer(c *cli.Context, config *ServerConfig) {
	if c.String("tcp") != "" {
		config.TCP = c.String("tcp")
	}
	if c.String("udp") != "" {
		config.UDP = c.String("udp")
	}
}

func cmdServer(c *cli.Context) error {
	/*
		if err := agent.Listen(&agent.Options{NoShutdownCleanup: true}); err != nil {
			log.Fatal(err)
		}
	*/

	config := c.App.Metadata["config"].(*ServerConfig)
	optOverwiteServer(c, config)

	wg := new(sync.WaitGroup)
	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)

	//
	// Setup
	//
	whoson.NewMainStore()
	whoson.NewLogger("stdout", "error")
	//whoson.NewLogger("stdout", "debug")
	//whoson.NewLogger("/tmp/gowhoson.log", "debug")
	//whoson.NewLogger("discard", "error")
	whoson.NewIDGenerator(uint32(1))

	var serverCount = 0
	var con *net.UDPConn
	if config.UDP != "nostart" {
		host, port, err := splitHostPort(config.UDP)
		if err != nil {
			displayError(c.App.ErrWriter, err)
			return err
		}
		addrudp := net.UDPAddr{
			Port: port,
			IP:   net.ParseIP(host),
		}
		con, err = net.ListenUDP("udp", &addrudp)
		if err != nil {
			displayError(c.App.ErrWriter, err)
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			whoson.ServeUDP(con)
		}()
		serverCount++
	}

	var lis *net.TCPListener
	if config.TCP != "nostart" {
		host, port, err := splitHostPort(config.TCP)
		if err != nil {
			displayError(c.App.ErrWriter, err)
			return err
		}
		addrtcp := net.TCPAddr{
			Port: port,
			IP:   net.ParseIP(host),
		}
		lis, err = net.ListenTCP("tcp", &addrtcp)
		if err != nil {
			displayError(c.App.ErrWriter, err)
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			whoson.ServeTCP(lis)
		}()
		serverCount++
	}

	lishttp, err := net.Listen("tcp", ":8080")
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		http.Serve(lishttp, nil)
	}()

	if serverCount > 0 {
		wg.Add(1)
		signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		go signalHandler(sigChan, wg, c, func() {
			if config.UDP != "nostart" {
				con.Close()
			}
			if config.TCP != "nostart" {
				lis.Close()
			}
			lishttp.Close()
		})
	}

	wg.Wait()
	return nil
}
