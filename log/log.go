package log

import (
	"io"
	stdlog "log"
	"os"

	"github.com/briandowns/sky-island/config"
	"github.com/go-kit/kit/log"
)

// stdlibAdapter
type stdlibAdapter struct {
	log.Logger
}

// newStdlibAdapter
func newStdlibAdapter(logger log.Logger) io.Writer {
	return stdlibAdapter{
		Logger: logger,
	}
}

func (a stdlibAdapter) Write(p []byte) (int, error) {
	if err := a.Logger.Log("msg", string(p)); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Logger
func Logger(conf *config.Config, name string) (log.Logger, error) {
	var logger log.Logger
	if conf.LogFile != "" {
		f, err := os.Open(conf.LogFile)
		if err != nil {
			return nil, err
		}
		mw := io.MultiWriter(os.Stdout, f)
		logger = log.NewJSONLogger(mw)
	} else {
		logger = log.NewJSONLogger(os.Stdout)
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "service", name)
	stdlog.SetOutput(newStdlibAdapter(logger))
	stdlog.SetFlags(0)
	return logger, nil
}
