package main

import (
	"context"
	"fmt"
	plgtran "github.com/lzkking/edr/plugins/lib"
	"sync"
	"time"
)

func handleRecv(ctx context.Context, wg *sync.WaitGroup, client *plgtran.Client) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t, err := client.ReceiveTask()
			if err != nil {
				return
			}

			fmt.Println(t)
		}
	}
}

func handleSend(ctx context.Context, wg *sync.WaitGroup, client *plgtran.Client) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		}
	}
}

func main() {
	plgClient := plgtran.New()

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Add(2)

	go handleSend(ctx, wg, plgClient)

	go func() {
		handleRecv(ctx, wg, plgClient)
		cancel()
	}()

	wg.Wait()
}
