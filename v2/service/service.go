package service

import (
	"context"
	"time"

	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

// Start starts a process.
func Start(ctx context.Context, logger log.Logger, svc, uid string) (err error) {
	logger = logger.Add("method", "Start", "service", svc)
	for {
		select {
		case <-ctx.Done():
			if err != nil {
				return err
			}
			return ctx.Err()
		default:
			if !util.IsServiceStopped(svc, uid) {
				logger.Info("service is running")
				return nil
			}
			logger.Info("attempting to start service")
			err = startService(svc, uid)
			if err != nil {
				logger.Warn("start service error: " + err.Error())
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Stop stops tor process.
func Stop(ctx context.Context, logger log.Logger, svc, uid string) (err error) {
	logger = logger.Add("method", "Stop", "service", svc)
	for {
		select {
		case <-ctx.Done():
			if err != nil {
				return err
			}
			return ctx.Err()
		default:
			if util.IsServiceStopped(svc, uid) {
				logger.Info("service stopped")
				return nil
			}
			logger.Info("stop service " + svc)
			err = stopService(svc, uid)
			if err != nil {
				logger.Warn("stop service error: " + err.Error())
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
