package bench

import (
	"context"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"log"
	"math/big"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"

	"github.com/LampardNguyen234/evm-latency-bench/pkg/bench"
)

var ReceiptCountCommand = &cli.Command{
	Name:  "receiptcount",
	Usage: "Send transactions and count eth_getTransactionReceipt calls per transaction",
	Flags: BenchFlags,
	Action: func(c *cli.Context) error {
		if err := bench.LoadEnv(c.String("env-file")); err != nil {
			return fmt.Errorf("failed to load env: %w", err)
		}

		rpcEndpoint := bench.RPCEndpoint()
		client, err := ethclient.Dial(rpcEndpoint)
		if err != nil {
			return fmt.Errorf("failed to connect RPC endpoint: %w", err)
		}
		defer client.Close()

		txCount := c.Int("txcount")
		pollInterval := c.Duration("poll-interval")
		plotEnabled := c.Bool("plot")
		plotDir := c.String("plot-dir")
		plotPrefix := c.String("plot-prefix")
		ctx := context.Background()

		fmt.Printf("Sending %d transactions and counting receipt polling calls...\n", txCount)

		receiptCallCounts := make([]int, 0, txCount)

		for i := 0; i < txCount; i++ {
			keyHex := bench.PrivKeys()[i%len(bench.PrivKeys())]
			privKey, err := crypto.HexToECDSA(keyHex)
			if err != nil {
				return fmt.Errorf("invalid private key: %w", err)
			}
			fromAddress := crypto.PubkeyToAddress(privKey.PublicKey)

			nonce, err := client.PendingNonceAt(ctx, fromAddress)
			if err != nil {
				return fmt.Errorf("failed to get nonce: %w", err)
			}

			toAddress := fromAddress  // self-transfer
			value := big.NewInt(1e10) // 0.00000000001 ETH
			gasLimit := uint64(21000)

			chainID, err := client.NetworkID(ctx)
			if err != nil {
				return fmt.Errorf("failed to get network ID: %w", err)
			}

			gasTipCap, err := client.SuggestGasTipCap(ctx)
			if err != nil {
				return fmt.Errorf("failed to get gas tip cap: %w", err)
			}

			gasFeeCap, err := client.SuggestGasPrice(ctx)
			if err != nil {
				return fmt.Errorf("failed to get gas fee cap: %w", err)
			}

			txData := &types.DynamicFeeTx{
				ChainID:   chainID,
				Nonce:     nonce,
				GasTipCap: gasTipCap,
				GasFeeCap: gasFeeCap,
				Gas:       gasLimit,
				To:        &toAddress,
				Value:     value,
				Data:      nil,
			}

			tx := types.NewTx(txData)

			signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), privKey)
			if err != nil {
				return fmt.Errorf("failed to sign transaction: %w", err)
			}

			err = client.SendTransaction(ctx, signedTx)
			if err != nil {
				return fmt.Errorf("failed to send transaction: %w", err)
			}

			txHash := signedTx.Hash()
			log.Printf("[INFO] Tx %d sent: %s", i+1, txHash.Hex())

			// Poll for receipt and count calls
			receiptCallCount := 0
			var receipt *types.Receipt
			for {
				receiptCallCount++
				receipt, err = client.TransactionReceipt(ctx, txHash)
				if err == nil && receipt != nil {
					break
				}
				time.Sleep(pollInterval)
			}

			log.Printf("[INFO] Tx %d receipt obtained after %d eth_getTransactionReceipt calls", i+1, receiptCallCount)
			fmt.Printf("Tx %d: Receipt calls = %d\n", i+1, receiptCallCount)

			receiptCallCounts = append(receiptCallCounts, receiptCallCount)
		}

		if plotEnabled {
			plotFile := filepath.Join(plotDir, plotPrefix+".png")
			if err := plotReceiptCallCounts(receiptCallCounts, plotFile); err != nil {
				fmt.Printf("Warning: failed to generate receipt call count plot: %v\n", err)
			} else {
				fmt.Printf("Receipt call count plot saved as '%s'\n", plotFile)
			}
		}

		return nil
	},
}

// plotReceiptCallCounts plots the receipt call counts per transaction as a line chart.
func plotReceiptCallCounts(counts []int, filename string) error {
	pts := make(plotter.XYs, len(counts))
	for i, c := range counts {
		pts[i].X = float64(i + 1)
		pts[i].Y = float64(c)
	}

	p := plot.New()
	p.Title.Text = "eth_getTransactionReceipt Calls Per Transaction"
	p.X.Label.Text = "Transaction #"
	p.Y.Label.Text = "Receipt Call Count"
	p.Legend.Top = true
	p.Legend.Left = false
	p.Add(plotter.NewGrid())

	err := plotutil.AddLinePoints(p, "Receipt Calls", pts)
	if err != nil {
		return err
	}

	if err := p.Save(6*vg.Inch, 4*vg.Inch, filename); err != nil {
		return err
	}
	return nil
}
