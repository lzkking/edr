package main

import (
	"context"
	"github.com/lzkking/edr/agent/internal/transfer"
	"sync"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	wg.Add(1)

	go transfer.Transfer(ctx, &wg)

	wg.Wait()

	cancel()

}
