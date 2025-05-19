package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/LampardNguyen234/evm-latency-bench/cmd/bench"
)

func main() {
	app := &cli.App{
		Name:  "evmbench",
		Usage: "EVM benchmarking CLI tool",
		Commands: []*cli.Command{
			bench.BenchCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
