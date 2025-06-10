package main

import (
	"context"
	"github.com/alecthomas/kong"
	"github.com/sz-po/go-distributed-kvm-switch/internal/app/kernel"
	"log/slog"
	"os"
	"sync"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx := context.Background()
	wg := &sync.WaitGroup{}

	config := kernel.Config{}

	if result := kong.Parse(&config); result.Error != nil {
		slog.Error("Failed to parse config.", "error", result.Error)
	}

	if err := kernel.Start(ctx, wg, config); err != nil {
		panic(err)
	}

	slog.Info("Kernel started.")
	wg.Wait()

	os.Exit(0)
}
