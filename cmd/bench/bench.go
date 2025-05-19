package bench

import (
	"fmt"
	"github.com/LampardNguyen234/evm-latency-bench/pkg/bench"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"path/filepath"
	"time"
)

var BenchFlags = []cli.Flag{
	&cli.IntFlag{
		Name:    "txcount",
		Aliases: []string{"n"},
		Usage:   "Number of transactions to send sequentially",
		Value:   10,
	},
	&cli.DurationFlag{
		Name:    "poll-interval",
		Usage:   "Polling interval for receipt queries (only async mode)",
		Value:   1 * time.Second,
		EnvVars: []string{"POLL_INTERVAL_MS"},
	},
	&cli.StringFlag{
		Name:  "env-file",
		Usage: "Path to .env file with RPC_ENDPOINT and PRIVATE_KEYS",
		Value: ".env",
	},
	&cli.StringFlag{
		Name:  "mode",
		Usage: "Transaction submission mode: 'async' or 'sync'",
		Value: "async",
	},
	&cli.BoolFlag{
		Name:  "plot",
		Usage: "Generate PNG plots for the benchmark results",
		Value: false,
	},
	&cli.StringFlag{
		Name:  "plot-prefix",
		Usage: "Filename prefix for output PNG plots",
		Value: "benchmark_results",
	},
	&cli.StringFlag{
		Name:  "plot-dir",
		Usage: "Directory to save PNG plot files",
		Value: ".",
	},
}

var BenchCommand = &cli.Command{
	Name:        "bench",
	Usage:       "Benchmark EVM transaction submission and receipt latency",
	Flags:       BenchFlags,
	Subcommands: []*cli.Command{CompareSubcommand, ReceiptCountCommand, BlockNumberCommand},
	Action: func(c *cli.Context) error {
		envFile := c.String("env-file")
		if err := bench.LoadEnv(envFile); err != nil {
			return err
		}

		txCount := c.Int("txcount")
		pollInterval := c.Duration("poll-interval")
		mode := c.String("mode")
		plotEnabled := c.Bool("plot")
		plotPrefix := c.String("plot-prefix")
		plotDir := c.String("plot-dir")

		var results []bench.Result
		var err error

		fmt.Println("Extracting RPC response time metrics...")
		client, err := ethclient.Dial(bench.RPCEndpoint())
		if err != nil {
			return fmt.Errorf("failed to connect RPC endpoint: %w", err)
		}
		defer client.Close()

		metrics := extractRPCTime(client, 10, 100*time.Millisecond)
		if metrics == nil {
			return fmt.Errorf("failed to extract RPC response time metrics")
		}

		switch mode {
		case "async":
			results, err = bench.RunBenchmarkAsync(txCount, pollInterval)
		case "sync":
			results, err = bench.RunBenchmarkSync(txCount)
		default:
			return fmt.Errorf("invalid mode: %s, must be 'async' or 'sync'", mode)
		}
		if err != nil {
			return err
		}

		bench.PrintReport(results)

		if plotEnabled {
			fullPath := filepath.Join(plotDir, plotPrefix+"_combined.png")
			if err := bench.PlotCombinedMetrics(results, metrics[3], fullPath); err != nil {
				fmt.Printf("Warning: failed to generate combined plot: %v\n", err)
			} else {
				fmt.Printf("Combined benchmark plot saved as '%s'\n", fullPath)
			}
		}

		return nil
	},
}
