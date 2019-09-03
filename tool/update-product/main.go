package main

import (
	"context"
	"flag"

	"github.com/privatix/dapp-installer/v2/product"
	"github.com/privatix/dappctrl/util/log"
)

func main() {
	old := flag.String("old", "./client.old", "Old install direcotry")
	newd := flag.String("new", "./client", "New install directory")
	role := flag.String("role", "client", "client | agent")

	flag.Parse()

	l, _ := log.NewStderrLogger(log.NewWriterConfig())
	logger := log.NewMultiLogger(l)

	if err := product.UpdateAll(context.Background(), logger, *old, *newd, *role); err != nil {
		logger.Fatal(err.Error())
	}
	logger.Info("done.")
}
