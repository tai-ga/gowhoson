package gowhoson

import (
	"bytes"
	"context"
	"encoding/json"
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

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/tai-ga/gowhoson/pkg/whoson"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func signalHandler(ctx context.Context, ch <-chan os.Signal, wg *sync.WaitGroup, c *cli.Command, f func()) {
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
					displayError(c.Root().ErrWriter, errors.New("clean shutdown took too long, forcing exit"))
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

func ipportsValidate(c *cli.Command, optname string) (string, error) {
	ipports := c.String(optname)
	ipportList := strings.Split(ipports, ",")

	for i, ipport := range ipportList {
		ipport = strings.TrimSpace(ipport)
		host, _, err := splitHostPort(ipport)
		if err != nil {
			return "", err
		}
		if net.ParseIP(host) == nil {
			return "", fmt.Errorf("\"--%s %s\" parse error", optname, ipports)
		}
		ipportList[i] = ipport
	}

	return strings.Join(ipportList, ","), nil
}

func cmdServerValidate(c *cli.Command) (*whoson.ServerConfig, error) {
	config := c.Root().Metadata["config"].(*whoson.ServerConfig)
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

	config.Expvar = c.Bool("expvar")

	validatePortOption := func(optName string) error {
		optValue := c.String(optName)
		if optValue != "" && optValue != "nostart" {
			ipports, err := ipportsValidate(c, optName)
			if err != nil {
				return err
			}

			switch optName {
			case "tcp":
				config.TCP = ipports
			case "udp":
				config.UDP = ipports
			case "controlport":
				config.ControlPort = ipports
			case "syncremote":
				config.SyncRemote = ipports
			}
		}
		return nil
	}

	for _, opt := range []string{"tcp", "udp", "controlport", "syncremote"} {
		if err := validatePortOption(opt); err != nil {
			return nil, err
		}
	}

	if c.String("savefile") != "" {
		config.SaveFile = c.String("savefile")
	}
	return config, nil
}

func cmdServer(ctx context.Context, c *cli.Command) error {
	var err error
	config, err := cmdServerValidate(c)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return err
	}

	wg := new(sync.WaitGroup)
	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)

	whoson.NewMainStoreEnableSyncRemote()
	err = loadStore(config.SaveFile)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return err
	}
	err = whoson.NewLogger(config.Log, config.Loglevel)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return err
	}
	whoson.Log("info", fmt.Sprintf("ServerID:%d", config.ServerID), nil, nil)
	whoson.NewIDGenerator(uint(config.ServerID))

	var con *net.UDPConn
	if config.UDP != "nostart" {
		con, err = runUDPServer(c, config, wg)
		if err != nil {
			return err
		}
	}

	var lis *net.TCPListener
	if config.TCP != "nostart" {
		lis, err = runTCPServer(c, config, wg)
		if err != nil {
			return err
		}
	}

	var lishttp net.Listener
	lishttp, err = runExpvar(config, wg, c)
	if err != nil {
		return err
	}

	var g *grpc.Server
	var lisgrpc net.Listener

	// Initialize logging interceptor
	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	zapLogger := &zapLoggerAdapter{logger: whoson.Logger}

	g = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(zapLogger, logOpts...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(zapLogger, logOpts...),
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
		if config.SyncRemote != "" {
			hosts := strings.Split(config.SyncRemote, ",")
			whoson.RunSyncRemote(ctx, hosts)
		}
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
			displayError(c.Root().ErrWriter, err)
		}
	})

	wg.Wait()
	return nil
}

func runUDPServer(c *cli.Command, config *whoson.ServerConfig, wg *sync.WaitGroup) (*net.UDPConn, error) {
	host, port, err := splitHostPort(config.UDP)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return nil, err
	}
	addrudp := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}
	var con *net.UDPConn
	con, err = net.ListenUDP("udp", &addrudp)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return nil, err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		whoson.ServeUDP(con)
	}()
	return con, nil
}

func runTCPServer(c *cli.Command, config *whoson.ServerConfig, wg *sync.WaitGroup) (*net.TCPListener, error) {
	var lis *net.TCPListener
	host, port, err := splitHostPort(config.TCP)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return nil, err
	}
	addrtcp := net.TCPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}
	lis, err = net.ListenTCP("tcp", &addrtcp)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return nil, err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		whoson.ServeTCP(lis)
	}()
	return lis, nil
}

func getListener(c *cli.Command, host string) (net.Listener, error) {
	l, err := net.Listen("tcp", host)
	if err != nil {
		displayError(c.Root().ErrWriter, err)
		return nil, err
	}
	return l, nil
}

func runExpvar(config *whoson.ServerConfig, wg *sync.WaitGroup, c *cli.Command) (net.Listener, error) {
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

func runGrpc(g *grpc.Server, config *whoson.ServerConfig, wg *sync.WaitGroup, c *cli.Command) (net.Listener, error) {
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
	b, err := os.ReadFile(f)
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
		err = os.WriteFile(f, []byte(""), 0644)
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

	err = os.WriteFile(f, b.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

type zapLoggerAdapter struct {
	logger *zap.Logger
}

func (l *zapLoggerAdapter) Log(ctx context.Context, level logging.Level, msg string, fields ...any) {
	zapLevel := zapcore.InfoLevel
	switch level {
	case logging.LevelDebug:
		zapLevel = zapcore.DebugLevel
	case logging.LevelInfo:
		zapLevel = zapcore.InfoLevel
	case logging.LevelWarn:
		zapLevel = zapcore.WarnLevel
	case logging.LevelError:
		zapLevel = zapcore.ErrorLevel
	}

	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}

	l.logger.Log(zapLevel, msg, zapFields...)
}
