package bench

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/LampardNguyen234/evm-latency-bench/pkg/bench"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
)

var BlockNumberCommand = &cli.Command{
	Name:  "resp-time",
	Usage: "Measure time to call eth_blockNumber RPC",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "env-file",
			Usage: "Path to .env file with RPC_ENDPOINT",
			Value: ".env",
		},
		&cli.IntFlag{
			Name:  "count",
			Usage: "Number of times to call eth_blockNumber",
			Value: 10,
		},
		&cli.DurationFlag{
			Name:  "interval",
			Usage: "Interval between calls",
			Value: 500 * time.Millisecond,
		},
	},
	Action: func(c *cli.Context) error {
		if err := bench.LoadEnv(c.String("env-file")); err != nil {
			return fmt.Errorf("failed to load env: %w", err)
		}

		client, err := ethclient.Dial(bench.RPCEndpoint())
		if err != nil {
			return fmt.Errorf("failed to connect RPC endpoint: %w", err)
		}
		defer client.Close()

		count := c.Int("count")
		interval := c.Duration("interval")

		fmt.Printf("Calling eth_blockNumber %d times with %v interval...\n", count, interval)

		times := make([]time.Duration, 0, count)

		ctx := context.Background()

		for i := 0; i < count; i++ {
			start := time.Now()
			_, err := client.BlockNumber(ctx)
			elapsed := time.Since(start)

			if err != nil {
				log.Printf("Call %d failed: %v", i+1, err)
			} else {
				fmt.Printf("Call %d took %v\n", i+1, elapsed)
				times = append(times, elapsed)
			}

			if i < count-1 {
				time.Sleep(interval)
			}
		}

		if len(times) == 0 {
			fmt.Println("No successful calls to measure.")
			return nil
		}

		// Calculate minTime, maxTime, avg, median
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

		minTime := times[0]
		maxTime := times[len(times)-1]

		var total time.Duration
		for _, t := range times {
			total += t
		}
		avg := total / time.Duration(len(times))

		median := times[len(times)/2]
		if len(times)%2 == 0 {
			median = (times[len(times)/2-1] + times[len(times)/2]) / 2
		}

		fmt.Println("\neth_blockNumber call time statistics:")
		fmt.Printf("Min:    %v\n", minTime)
		fmt.Printf("Max:    %v\n", maxTime)
		fmt.Printf("Avg:    %v\n", avg)
		fmt.Printf("Median: %v\n", median)

		return nil
	},
}
