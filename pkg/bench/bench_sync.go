package bench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func getChainID(ctx context.Context, client *ethclient.Client) (*big.Int, error) {
	return client.NetworkID(ctx)
}

func sendRawTransactionSyncWithMethod(rpcURL, rawTxHex, method string) (json.RawMessage, error) {
	req := rpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  []interface{}{rawTxHex},
		ID:      1,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal RPC request failed: %w", err)
	}

	resp, err := http.Post(rpcURL, "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("RPC HTTP POST failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read RPC response failed: %w", err)
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshal RPC response failed: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func RunBenchmarkSync(txCount int) ([]Result, error) {
	client, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC endpoint: %w", err)
	}
	ctx := context.Background()

	chainID, err := getChainID(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Detect special chain ID "6342"
	specialChainID := big.NewInt(6342)
	useRealtimeMethod := chainID.Cmp(specialChainID) == 0

	var rpcMethod string
	if useRealtimeMethod {
		rpcMethod = "realtime_sendRawTransaction"
	} else {
		rpcMethod = "eth_sendRawTransactionSync"
	}

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
		toAddress := fromAddress  // self-transfer
		value := big.NewInt(1e10) // 0.00000000001 ETH
		gasLimit := uint64(21000)

		gasTipCap, err := client.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas tip cap: %w", err)
		}

		gasFeeCap, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas fee cap: %w", err)
		}
		gasFeeCap = gasFeeCap.Mul(gasFeeCap, big.NewInt(2))

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

		// Use LondonSigner for EIP-1559 transactions
		signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), privKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %w", err)
		}

		rawTxBytes, err := signedTx.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal signed tx: %w", err)
		}
		rawTxHex := "0x" + fmt.Sprintf("%x", rawTxBytes)

		sendStart := time.Now()
		resultRaw, err := sendRawTransactionSyncWithMethod(rpcEndpoint, rawTxHex, rpcMethod)
		sendEnd := time.Now()
		if err != nil {
			// Log the error and continue with next transaction
			log.Printf("[WARN] Tx %d: RPC call failed: %v. Skipping and continuing.", i+1, err)
			time.Sleep(2 * time.Second)
		} else {
			sendDuration := sendEnd.Sub(sendStart)

			var receipt types.Receipt
			if err := json.Unmarshal(resultRaw, &receipt); err != nil {
				// Log but continue
				log.Printf("[WARN] Tx %d: failed to unmarshal receipt: %v", i+1, err)
			} else {
				log.Printf("[INFO] Tx %d: receipt status: %d, block number: %d", i+1, receipt.Status, receipt.BlockNumber.Uint64())
			}

			txHash := signedTx.Hash()
			log.Printf("[INFO] Tx %d: sent and received receipt for %s in %v", i+1, txHash.Hex(), sendDuration)

			results = append(results, Result{
				TxIndex:     i + 1,
				TxHash:      txHash.Hex(),
				SendTime:    sendDuration.Milliseconds(),
				ConfirmTime: 0,
				TotalTime:   sendDuration.Milliseconds(),
			})
		}

		nonce++
	}

	return results, nil
}
