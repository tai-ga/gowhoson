package whoson

import (
	"errors"
	"time"

	"github.com/client9/reopen"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger return new zap.Logger struct pointer.
func NewLogger(output, loglevel string) *zap.Logger {
	if Logger == nil {
		InitLog(output, loglevel)
	}
	return Logger
}

// InitLog initial setup for Logger.
func InitLog(output, loglevel string) error {
	var writer reopen.Writer
	switch output {
	case "stdout":
		writer = reopen.Stdout
	case "stderr":
		writer = reopen.Stderr
	case "discard":
		writer = reopen.Discard
	default:
		f, err := reopen.NewFileWriterMode(output, 0644)
		if err != nil {
			return err
		}
		writer = f
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(loglevel)); err != nil {
		return err
	}

	config := zap.NewProductionConfig().EncoderConfig
	config.TimeKey = "time"
	config.MessageKey = "msg"
	config.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006/01/02 15:04:05 MST"))
	}

	encoder := zapcore.NewJSONEncoder(config)
	writeSyncer := zapcore.AddSync(writer)
	Logger = zap.New(
		zapcore.NewCore(
			encoder,
			zapcore.Lock(writeSyncer),
			level,
		),
		zap.ErrorOutput(writeSyncer),
	)

	return nil
}

func switchLogger(status string) (func(string, ...zapcore.Field), error) {
	var logger func(string, ...zapcore.Field)

	switch status {
	case "debug":
		logger = Logger.Debug
	case "info":
		logger = Logger.Info
	case "warn":
		logger = Logger.Warn
	case "error":
		logger = Logger.Error
	case "dpanic":
		logger = Logger.DPanic
	case "panic":
		logger = Logger.Panic
	case "fatal":
		logger = Logger.Fatal
	default:
		return nil, errors.New("Log status not found")
	}
	return logger, nil
}

// Log is API entry point of Logging.
func Log(status, msg string, ses *Session, err error) {
	logger, err := switchLogger(status)
	if err != nil {
		return
	}

	id := zap.Skip()
	protocol := zap.Skip()
	remote := zap.Skip()
	cmdMethod := zap.Skip()
	cmdIP := zap.Skip()
	cmdArgs := zap.Skip()
	if ses != nil && ses.id != 0 {
		id = zap.Uint64("id", ses.id)
		if ses.protocol == pTCP {
			protocol = zap.String("protocol", "tcp")
			remote = zap.String("remote", ses.conn.RemoteAddr().String())
		} else if ses.protocol == pUDP {
			protocol = zap.String("protocol", "udp")
			remote = zap.String("remote", ses.remoteAddr.String())
		}
		if method[ses.cmdMethod] != "" {
			cmdMethod = zap.String("cmd", method[ses.cmdMethod])
		}
		if ses.cmdIP != nil {
			cmdIP = zap.String("cmdip", ses.cmdIP.String())
		}
		if ses.cmdArgs != "" {
			cmdArgs = zap.String("cmdargs", ses.cmdArgs)
		}
	}
	errMsg := zap.Skip()
	if err != nil {
		errMsg = zap.String("error", err.Error())
	}

	logger(msg,
		id,
		protocol,
		remote,
		cmdMethod,
		cmdIP,
		cmdArgs,
		errMsg,
	)
}
