package util

import (
	"io"

	"github.com/privatix/dappctrl/util/log"
)

// CreateLogger creates new logger file.
func CreateLogger() (log.Logger, io.Closer, error) {
	logConfig := &log.FileConfig{
		WriterConfig: log.NewWriterConfig(),
		Filename:     "dapp-installer-%Y-%m-%d.log",
		FileMode:     0644,
	}

	logger, closer, err := log.NewFileLogger(logConfig)
	if err != nil {
		return nil, nil, err
	}

	return logger, closer, nil
}
