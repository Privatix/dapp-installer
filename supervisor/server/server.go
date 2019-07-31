package server

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/privatix/dapp-installer/container"
	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/product"
	"github.com/privatix/dappctrl/util/log"
)

var actionTimeout = 10 * time.Second

func newContextWithTimeout() (context.Context, func()) {
	return context.WithTimeout(context.Background(), actionTimeout)
}

func writeErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

// ListenAndServe start listening for supervisor requests.
func ListenAndServe(logger log.Logger, addr string, d *dapp.Dapp, installUID string) error {
	var mu sync.Mutex

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		logger.Info("UID: " + installUID)
		logger.Info("stopping services...")
		if runtime.GOOS == "linux" {
			c := container.NewContainer()
			c.Name = d.Role
			c.Path = d.Path

			if err := c.Stop(); err != nil {
				logger.Error(err.Error())
				writeErr(w, err)
				return
			}
			logger.Info("done stopping services.")
			return
		}
		logger.Info("stopping services...")
		ctx, cancel := newContextWithTimeout()
		defer cancel()
		logger.Info(fmt.Sprintf("stopping service %s...", d.Tor.ServiceName()))
		err := d.Tor.Stop(ctx, installUID)
		if err != nil {
			err = fmt.Errorf("could not stop tor `%s`: %v", d.Tor.ServiceName(), err)
			logger.Error(err.Error())
			writeErr(w, err)
			return
		}
		// Defer recovery from further errors.
		defer func() {
			if err != nil {
				ctx, cancel := newContextWithTimeout()
				defer cancel()
				logger.Info(fmt.Sprintf("recovering %s...", d.Tor.ServiceName()))
				if err := d.Tor.Start(ctx, installUID); err != nil {
					logger.Error(err.Error())
				}
			}
		}()

		logger.Info("stopping products...")
		// Stop product services.
		err = product.Stop(d.Role, d.Path, d.Product)
		if err != nil {
			writeErr(w, fmt.Errorf("could not stop products: %v", err))
			return
		}
		// Defer recovery from further errors.
		defer func() {
			if err != nil {
				logger.Info("recovering products...")
				if err := product.Start(d.Role, d.Path, d.Product); err != nil {
					logger.Error(err.Error())
				}
			}
		}()

		logger.Info("stopping dappctrl and db...")
		// Stop dappctrl and db.
		ctx2, cancel2 := newContextWithTimeout()
		defer cancel2()
		err = d.Stop(ctx2, installUID)
		if err != nil {
			err = fmt.Errorf("could not stop dappctrl and db: timeout")
			logger.Error(err.Error())
			writeErr(w, err)
			// Recovering. Could not stop all services but might stop some.
			// Start ensures everything is back to running.
			logger.Info("recovering dappctrl and db...")
			ctx, cancel := newContextWithTimeout()
			defer cancel()
			if err := d.Start(ctx, installUID); err != nil {
				logger.Error(err.Error())
			}
			return
		}
		logger.Info("done stopping services.")
	})

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		logger.Info("UID: " + installUID)
		logger.Info("starting services...")
		if runtime.GOOS == "linux" {
			c := container.NewContainer()
			c.Name = d.Role
			c.Path = d.Path

			if err := c.Start(); err != nil {
				logger.Error(err.Error())
				writeErr(w, err)
				return
			}
			logger.Info("done starting services.")
			return
		}
		ctx, cancel := newContextWithTimeout()
		defer cancel()
		logger.Info(fmt.Sprintf("starting %s...", d.Tor.ServiceName()))
		logger.Info(d.Tor.RootPath)
		err := d.Tor.Start(ctx, installUID)
		if err != nil {
			err = fmt.Errorf("could not start tor: %v", err)
			logger.Error(err.Error())
			writeErr(w, err)
			return
		}
		// Defer recovery from further errors.
		defer func() {
			if err != nil {
				logger.Info(fmt.Sprintf("recovering %s...", d.Tor.ServiceName()))
				ctx, cancel := newContextWithTimeout()
				defer cancel()
				if err := d.Tor.Stop(ctx, installUID); err != nil {
					logger.Error(err.Error())
				}
			}
		}()

		logger.Info("starting dappctrl and db...")
		logger.Info(fmt.Sprint("controller:", d.Controller.Service.ID, " db:", dbengine.Hash(d.Path)))
		// Start dappctrl and db.
		ctx2, cancel2 := newContextWithTimeout()
		defer cancel2()
		err = d.Start(ctx2, installUID)
		if err != nil {
			err = fmt.Errorf("could not start dappctrl and db: %v", err)
			logger.Error(err.Error())
			writeErr(w, err)
			// Recovering. Could not stop all services but might stop some.
			// Start ensures everything is back to running.
			ctx, cancel := newContextWithTimeout()
			defer cancel()
			if err := d.Stop(ctx, installUID); err != nil {
				logger.Error(err.Error())
			}
			return
		}
		// Defer recovery from further errors.
		defer func() {
			if err != nil {
				logger.Error("recovering dappctrl and db...")
				ctx, cancel := newContextWithTimeout()
				defer cancel()
				if err := d.Stop(ctx, installUID); err != nil {
					logger.Error(err.Error())
				}
			}
		}()

		logger.Info("starting products...")
		// Start product services.
		err = product.Start(d.Role, d.Path, d.Product)
		if err != nil {
			err = fmt.Errorf("could not start products: %v", err)
			logger.Error(err.Error())
			writeErr(w, err)
		}
		logger.Info("done starting services.")
	})

	return http.ListenAndServe(addr, nil)
}
