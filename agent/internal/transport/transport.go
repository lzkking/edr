package transport

import (
	"context"
	"sync"
)

func StartUp(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	subWg := &sync.WaitGroup{}
	defer subWg.Wait()

	subWg.Add(1)
	go func() {
		startTransfer(subCtx, subWg)
		cancel()
	}()
}
