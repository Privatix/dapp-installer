package util

import (
	"os"

	"github.com/privatix/dappctrl/util/log"
)

// CreateLogger creates new logger file.
func CreateLogger(logFile string) (*os.File, log.Logger, error) {
	file, err := os.OpenFile(
		logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}

	logger, err := log.NewFileLogger(log.NewFileConfig(), file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return file, logger, nil
}
