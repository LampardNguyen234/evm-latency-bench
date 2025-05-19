package bench

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	rpcEndpoint string
	privKeys    []string
)

type Result struct {
	TxIndex     int
	TxHash      string
	SendTime    int64 // milliseconds
	ConfirmTime int64 // milliseconds
	TotalTime   int64 // milliseconds
}

func LoadEnv(path string) error {
	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("failed to load env file: %w", err)
	}
	rpcEndpoint = os.Getenv("RPC_ENDPOINT")
	if rpcEndpoint == "" {
		return errors.New("RPC_ENDPOINT not set in env file")
	}
	keys := os.Getenv("PRIVATE_KEYS")
	if keys == "" {
		return errors.New("PRIVATE_KEYS not set in env file")
	}
	privKeys = strings.Split(keys, ",")
	for i := range privKeys {
		privKeys[i] = strings.TrimSpace(privKeys[i])
	}
	if len(privKeys) == 0 {
		return errors.New("no private keys found in PRIVATE_KEYS")
	}
	return nil
}

// RPCEndpoint returns the loaded RPC endpoint string
func RPCEndpoint() string {
	return rpcEndpoint
}

// PrivKeys returns the loaded private keys slice
func PrivKeys() []string {
	return privKeys
}
