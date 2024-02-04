package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/0xERR0R/blocky/config"
	"github.com/0xERR0R/blocky/evt"
	"github.com/0xERR0R/blocky/log"
	"github.com/0xERR0R/blocky/server"
	"github.com/0xERR0R/blocky/util"
	"github.com/spf13/cobra"
)

func newMultiServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "multiserve <...configs>",
		Args:  cobra.ArbitraryArgs,
		Short: "start multiple blocky DNS servers",
		RunE:  startMultiServer,
	}
}

func startMultiServer(_ *cobra.Command, args []string) error {
	printBanner()
	var terminationErr error

	fmt.Println(fmt.Sprintf("Starting %d servers...", len(args)))

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for _, configPath := range args {
		fmt.Println("loading config", configPath)
		cfg, err := config.LoadConfig(configPath, isConfigMandatory)
		if err != nil {
			return err
		}

		log.Configure(&cfg.Log)

		ctx, cancelFn := context.WithCancel(context.Background())
		defer cancelFn()

		srv, err := server.NewServer(ctx, cfg)
		if err != nil {
			return fmt.Errorf("can't start server: %w", err)
		}

		const errChanSize = 10
		errChan := make(chan error)

		fmt.Println("starting server", configPath)

		srv.Start(ctx, errChan)

		fmt.Println("server started")
	}

	fmt.Println("all servers started")

	evt.Bus().Publish(evt.ApplicationStarted, util.Version, util.BuildTime)

	select {}

	return terminationErr
}
