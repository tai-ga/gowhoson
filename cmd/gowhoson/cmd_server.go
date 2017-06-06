package gowhoson

import (
	"context"
	"errors"
	"fmt"
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

func optOverwiteServer(c *cli.Context, config *whoson.ServerConfig) {
	if c.String("tcp") != "" {
		config.TCP = c.String("tcp")
	}
	if c.String("udp") != "" {
		config.UDP = c.String("udp")
	}
	if c.String("log") != "" {
		config.Log = c.String("log")
	}
	if c.String("loglevel") != "" {
		config.Loglevel = c.String("loglevel")
	}
	if c.Int("serverid") != 0 {
		config.ServerID = c.Int("serverid")
	}
	if c.Bool("expvar") != false {
		config.Expvar = c.Bool("expvar")
	}
}

func cmdServer(c *cli.Context) error {
	/*
		if err := agent.Listen(&agent.Options{NoShutdownCleanup: true}); err != nil {
			log.Fatal(err)
		}
	*/

	config := c.App.Metadata["config"].(*whoson.ServerConfig)
	optOverwiteServer(c, config)

	wg := new(sync.WaitGroup)
	ctx, ctxCancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)

	whoson.NewMainStore()
	whoson.NewLogger(config.Log, config.Loglevel)
	whoson.Log("info", fmt.Sprintf("ServerID:%d", config.ServerID), nil, nil)
	whoson.NewIDGenerator(uint32(config.ServerID))

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

	wg.Add(1)
	go func() {
		defer wg.Done()
		whoson.RunExpireChecker(ctx)
	}()

	var lishttp net.Listener
	var err error
	if config.Expvar {
		lishttp, err = net.Listen("tcp", ":8080")
		if err != nil {
			displayError(c.App.ErrWriter, err)
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			http.Serve(lishttp, nil)
		}()
	}

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
			if config.Expvar {
				lishttp.Close()
			}
			ctxCancel()
		})
	}

	wg.Wait()
	return nil
}
