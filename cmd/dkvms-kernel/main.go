package main

import (
	"context"
	"github.com/sz-po/go-distributed-kvm-switch/internal/app/kernel"
	"os"
)

func main() {
	ctx := context.Background()

	if err := kernel.Start(ctx); err != nil {
		panic(err)
	}

	os.Exit(0)
}
