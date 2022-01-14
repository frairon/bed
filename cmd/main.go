package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/frairon/bed"
	"golang.org/x/sync/errgroup"
)

var (
	chip   = flag.String("chip", "gpiochip0", "chip to use")
	output = flag.String("output", "wake_times.json", "file to use for storing the data")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errg, ctx := errgroup.WithContext(ctx)

	wakeup, err := bed.NewWakeup(*chip, *output)
	if err != nil {
		log.Fatalf("error initializing wake up: %v", err)
	}
	server := bed.NewServer(wakeup)

	errg.Go(func() error {
		defer log.Printf("server done")
		return server.Run(ctx)
	})
	errg.Go(func() error {
		defer log.Printf("wakeup done")
		return wakeup.Run(ctx)
	})

	defer func() {
		if err := wakeup.Close(); err != nil {
			log.Printf("error closing wakeup: %v", err)
		}
	}()
	errg.Go(func() error {
		c := make(chan os.Signal, 1)
		// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
		// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
		signal.Notify(c, os.Interrupt)

		// Block until we receive our signal.
		select {
		case <-c:
		case <-ctx.Done():
		}
		cancel()
		return nil
	})

	if err := errg.Wait(); err != nil {
		log.Printf("error running bed: %v", err)
	}
}
