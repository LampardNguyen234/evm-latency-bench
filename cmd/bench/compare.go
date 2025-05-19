package bench

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"path/filepath"
	"sort"

	"github.com/LampardNguyen234/evm-latency-bench/pkg/bench"
)

var CompareSubcommand = &cli.Command{
	Name:  "compare",
	Usage: "Compare benchmark results between async and sync modes",
	Flags: BenchFlags,
	Action: func(c *cli.Context) error {
		envFile := c.String("env-file")
		if err := bench.LoadEnv(envFile); err != nil {
			return err
		}

		txCount := c.Int("txcount")
		pollInterval := c.Duration("poll-interval")
		plotEnabled := c.Bool("plot")
		plotDir := c.String("plot-dir")
		plotPrefix := c.String("plot-prefix")

		fmt.Println("Running async benchmark...")
		asyncResults, err := bench.RunBenchmarkAsync(txCount, pollInterval)
		if err != nil {
			return fmt.Errorf("async benchmark failed: %w", err)
		}

		fmt.Println("Running sync benchmark...")
		syncResults, err := bench.RunBenchmarkSync(txCount)
		if err != nil {
			return fmt.Errorf("sync benchmark failed: %w", err)
		}

		// Print side-by-side total time table
		fmt.Println("\nSide-by-Side Total Time Comparison (ms):")
		fmt.Printf("%-6s %-15s %-15s\n", "TX#", "Async Total", "Sync Total")
		for i := 0; i < txCount; i++ {
			fmt.Printf("%-6d %-15d %-15d\n",
				i+1,
				asyncResults[i].TotalTime,
				syncResults[i].TotalTime,
			)
		}

		// Print averages summary
		median := func(results []bench.Result) int64 {
			sorted := make([]bench.Result, len(results))
			copy(sorted, results)
			sort.Slice(sorted, func(i, j int) bool { return sorted[i].TotalTime < sorted[j].TotalTime })
			mid := len(sorted) / 2
			if len(sorted)%2 == 0 {
				return (sorted[mid-1].TotalTime + sorted[mid].TotalTime) / 2
			}
			return sorted[mid].TotalTime
		}

		avg := func(results []bench.Result) float64 {
			var sum int64
			for _, r := range results {
				sum += r.TotalTime
			}
			return float64(sum) / float64(len(results))
		}

		fmt.Printf("\nMedian Total Time (ms): Async = %v, Sync = %v\n", median(asyncResults), median(syncResults))
		fmt.Printf("Avg Total Time (ms): Async = %v, Sync = %v\n", avg(asyncResults), avg(syncResults))

		if plotEnabled {
			fullPath := filepath.Join(plotDir, plotPrefix+".png")
			if err := bench.PlotCombinedTotalTimeWithMedian(asyncResults, syncResults, fullPath); err != nil {
				fmt.Printf("Warning: failed to generate combined plot: %v\n", err)
			} else {
				fmt.Printf("Combined benchmark plot saved as '%s'\n", fullPath)
			}
		}

		return nil
	},
}
