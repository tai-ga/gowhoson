package gowhoson

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
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
		err := cmdServerValidateTCP(c, config)
		if err != nil {
			return nil, err
		}
	}
	if c.String("udp") != "" {
		err := cmdServerValidateUDP(c, config)
		if err != nil {
			return nil, err
		}
	}
	if c.String("savefile") != "" {
		config.SaveFile = c.String("savefile")
	}
	return config, nil
}

func cmdServerValidateTCP(c *cli.Context, config *whoson.ServerConfig) error {
	config.TCP = c.String("tcp")
	if config.TCP != "nostart" {
		host, _, err := splitHostPort(config.TCP)
		if err != nil {
			return err
		}
		if net.ParseIP(host) == nil {
			return fmt.Errorf("\"--tcp %s\" parse error", config.TCP)
		}
	}
	return nil
}

func cmdServerValidateUDP(c *cli.Context, config *whoson.ServerConfig) error {
	config.UDP = c.String("udp")
	if config.UDP != "nostart" {
		host, _, err := splitHostPort(config.UDP)
		if err != nil {
			return err
		}
		if net.ParseIP(host) == nil {
			return fmt.Errorf("\"--udp %s\" parse error", config.UDP)
		}
	}
	return nil
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
	err = loadStore(config.SaveFile)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return err
	}
	err = whoson.NewLogger(config.Log, config.Loglevel)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return err
	}
	whoson.Log("info", fmt.Sprintf("ServerID:%d", config.ServerID), nil, nil)
	whoson.NewIDGenerator(uint(config.ServerID))

	var con *net.UDPConn
	if config.UDP != "nostart" {
		con, err = runUDPServer(c, config, wg)
	}

	var lis *net.TCPListener
	if config.TCP != "nostart" {
		lis, err = runTCPServer(c, config, wg)
	}

	var lishttp net.Listener
	lishttp, err = runExpvar(config, wg, c)
	if err != nil {
		return err
	}

	var g *grpc.Server
	var lisgrpc net.Listener

	opts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ns", duration.Nanoseconds())
		}),
	}
	g = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_zap.UnaryServerInterceptor(whoson.Logger, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_zap.StreamServerInterceptor(whoson.Logger, opts...),
		),
	)
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

		err = saveStore(config.SaveFile)
		if err != nil {
			displayError(c.App.ErrWriter, err)
		}
	})

	wg.Wait()
	return nil
}

func runUDPServer(c *cli.Context, config *whoson.ServerConfig, wg *sync.WaitGroup) (*net.UDPConn, error) {
	host, port, err := splitHostPort(config.UDP)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return nil, err
	}
	addrudp := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}
	var con *net.UDPConn
	con, err = net.ListenUDP("udp", &addrudp)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return nil, err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		whoson.ServeUDP(con)
	}()
	return con, nil
}

func runTCPServer(c *cli.Context, config *whoson.ServerConfig, wg *sync.WaitGroup) (*net.TCPListener, error) {
	var lis *net.TCPListener
	host, port, err := splitHostPort(config.TCP)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return nil, err
	}
	addrtcp := net.TCPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}
	lis, err = net.ListenTCP("tcp", &addrtcp)
	if err != nil {
		displayError(c.App.ErrWriter, err)
		return nil, err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		whoson.ServeTCP(lis)
	}()
	return lis, nil
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
	if lisgrpc, err = getListener(c, config.ControlPort); err != nil {
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

func loadStore(f string) error {
	if f == "" {
		return nil
	}
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	var sds []*whoson.StoreData
	err = json.Unmarshal(b, &sds)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			return nil
		}
		return err
	}
	for _, sd := range sds {
		if sd.Expire.After(time.Now()) {
			whoson.MainStore.SyncSet(sd.IP.String(), sd)
		}
	}
	return nil
}

func saveStore(f string) error {
	if f == "" {
		return nil
	}
	jsonb, err := whoson.MainStore.ItemsJSON()
	if err != nil {
		return err
	}

	if string(jsonb) == "null" {
		err = ioutil.WriteFile(f, []byte(""), 0644)
		if err != nil {
			return err
		}
		return nil
	}

	var b bytes.Buffer
	err = json.Indent(&b, jsonb, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(f, b.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}
