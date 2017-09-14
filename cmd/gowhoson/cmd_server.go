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
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/tai-ga/gowhoson/whoson"
	"github.com/urfave/cli"
)

func signalHandler(ctx context.Context, ch <-chan os.Signal, wg *sync.WaitGroup, c *cli.Context, f func()) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case s := <-ch:
			switch s {
			case syscall.SIGHUP:
				err := whoson.LogWriter.Reopen()
				if err != nil {
					panic(err)
				}
			default:
				f()
				time.AfterFunc(time.Second*8, func() {
					displayError(c.App.ErrWriter, errors.New("Clean shutdown took too long, forcing exit"))
					os.Exit(0)
				})
			}
		}
	}
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

func cmdServerValidate(c *cli.Context) (*whoson.ServerConfig, error) {
	config := c.App.Metadata["config"].(*whoson.ServerConfig)
	if c.String("loglevel") != "" {
		config.Loglevel = c.String("loglevel")
		switch config.Loglevel {
		case "debug", "info", "warn", "error", "dpanic", "panic", "fatal":
		default:
			return nil, fmt.Errorf("\"--loglevel %s\" not support loglevel", config.Loglevel)
		}
	}
	if c.String("log") != "" {
		config.Log = c.String("log")
	}
	if c.Int("serverid") != 0 {
		config.ServerID = c.Int("serverid")
	}
	if c.Bool("expvar") != false {
		config.Expvar = c.Bool("expvar")
	}
	if c.String("tcp") != "" {
		config.TCP = c.String("tcp")
		if config.TCP != "nostart" {
			host, _, err := splitHostPort(config.TCP)
			if err != nil {
				return nil, err
			}
			if net.ParseIP(host) == nil {
				return nil, fmt.Errorf("\"--tcp %s\" parse error", config.TCP)
			}
		}
	}
	if c.String("udp") != "" {
		config.UDP = c.String("udp")
		if config.UDP != "nostart" {
			host, _, err := splitHostPort(config.UDP)
			if err != nil {
				return nil, err
			}
			if net.ParseIP(host) == nil {
				return nil, fmt.Errorf("\"--udp %s\" parse error", config.UDP)
			}
		}
	}
	return config, nil
}

func cmdServer(c *cli.Context) error {
	var err error
	config, err := cmdServerValidate(c)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return err
	}

	wg := new(sync.WaitGroup)
	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)

	whoson.NewMainStoreEnableSyncRemote()
	err = whoson.NewLogger(config.Log, config.Loglevel)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return err
	}
	whoson.Log("info", fmt.Sprintf("ServerID:%d", config.ServerID), nil, nil)
	whoson.NewIDGenerator(uint(config.ServerID))

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
	}

	var lishttp net.Listener
	lishttp, err = runExpvar(config, wg, c)
	if err != nil {
		return err
	}

	var g *grpc.Server
	var lisgrpc net.Listener
	g = grpc.NewServer()
	lisgrpc, err = runGrpc(g, config, wg, c)
	if err != nil {
		return err
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		whoson.RunExpireChecker(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		hosts := strings.Split(config.SyncRemote, ",")
		whoson.RunSyncRemote(ctx, hosts)
	}()

	wg.Add(1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go signalHandler(ctx, sigChan, wg, c, func() {
		defer ctxCancel()
		if config.UDP != "nostart" {
			con.Close()
		}
		if config.TCP != "nostart" {
			lis.Close()
		}
		if config.Expvar {
			lishttp.Close()
		}
		lisgrpc.Close()
		g.Stop()
	})

	wg.Wait()
	return nil
}

func getListener(c *cli.Context, host string) (net.Listener, error) {
	l, err := net.Listen("tcp", host)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return nil, err
	}
	return l, nil
}

func runExpvar(config *whoson.ServerConfig, wg *sync.WaitGroup, c *cli.Context) (net.Listener, error) {
	var lishttp net.Listener
	var err error
	if config.Expvar {
		if lishttp, err = getListener(c, ":8080"); err != nil {
			return nil, err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			http.Serve(lishttp, nil)
		}()
	}
	return lishttp, nil
}

func runGrpc(g *grpc.Server, config *whoson.ServerConfig, wg *sync.WaitGroup, c *cli.Context) (net.Listener, error) {
	var lisgrpc net.Listener
	var err error
	if lisgrpc, err = getListener(c, config.GRPCPort); err != nil {
		return nil, err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		whoson.RegisterSyncServer(g, &whoson.Sync{})
		g.Serve(lisgrpc)
	}()
	return lisgrpc, nil
}
