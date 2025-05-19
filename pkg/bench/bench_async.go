package bench

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func RunBenchmarkAsync(txCount int, pollInterval time.Duration) ([]Result, error) {
	client, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC endpoint: %w", err)
	}
	ctx := context.Background()

	results := make([]Result, 0, txCount)

	privKey, err := crypto.HexToECDSA(privKeys[0])
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	fromAddress := crypto.PubkeyToAddress(privKey.PublicKey)

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}
	for i := 0; i < txCount; i++ {
		time.Sleep(10 * time.Millisecond)

		log.Printf("[INFO] Tx %d: nonce %d from %s", i+1, nonce, fromAddress.Hex())

		toAddress := fromAddress  // self-transfer
		value := big.NewInt(1e10) // 0.001 ETH
		gasLimit := uint64(21000)
		gasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas price: %w", err)
		}

		tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
		chainID, err := client.NetworkID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get network ID: %w", err)
		}

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %w", err)
		}

		sendStart := time.Now()
		err = client.SendTransaction(ctx, signedTx)
		sendEnd := time.Now()
		if err != nil {
			return nil, fmt.Errorf("failed to send transaction: %w", err)
		}
		sendDuration := sendEnd.Sub(sendStart)

		txHash := signedTx.Hash()
		log.Printf("[INFO] Tx %d: sent %s in %v", i+1, txHash.Hex(), sendDuration)

		confirmStart := time.Now()
		var receipt *types.Receipt
		pollCount := 0
		for {
			receipt, err = client.TransactionReceipt(ctx, txHash)
			if err == nil && receipt != nil {
				break
			}
			pollCount++
			log.Printf("[DEBUG] Tx %d: polling receipt attempt %d", i+1, pollCount)
			time.Sleep(pollInterval)
		}
		confirmEnd := time.Now()
		confirmDuration := confirmEnd.Sub(confirmStart)

		log.Printf("[INFO] Tx %d: receipt confirmed in %v (polls: %d)\n\n", i+1, confirmDuration, pollCount)

		totalDuration := sendDuration + confirmDuration

		results = append(results, Result{
			TxIndex:     i + 1,
			TxHash:      txHash.Hex(),
			SendTime:    sendDuration.Milliseconds(),
			ConfirmTime: confirmDuration.Milliseconds(),
			TotalTime:   totalDuration.Milliseconds(),
		})
		nonce++
	}

	return results, nil
}
