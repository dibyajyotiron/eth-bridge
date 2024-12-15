package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// CurrencyConfig holds conversion factors and display names for different tokens
type CurrencyConfig struct {
	Factor   uint
	Currency string
}

type CurrencyConfigMap map[string]CurrencyConfig

type Config struct {
	PostgresURL     string
	RedisURL        string
	RedisStreamName string
	RedisStreamDlq  string
	EthereumRPCURL  string
	SocketGateAddr  string
	ServerPort      string
	ContractABI     string
	TopicHex        string

	// Define currency configurations
	CurrencyConfigs CurrencyConfigMap
}

func LoadConfig(envPath ...string) *Config {
	if len(envPath) != 0 && len(envPath) != 1 {
		log.Fatal("envPath provided is wrong")
	}
	if len(envPath) == 1 {
		err := godotenv.Load(envPath...)
		if err != nil {
			log.Printf(".env not found at %+v path", envPath[0])
		}
	}
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	return &Config{
		PostgresURL:     os.Getenv("DATABASE_URL"),
		RedisURL:        os.Getenv("REDIS_URL"),
		RedisStreamName: os.Getenv("REDIS_STREAM"),
		RedisStreamDlq:  os.Getenv("REDIS_STREAM_DLQ"),
		EthereumRPCURL:  os.Getenv("ETHEREUM_RPC_URL"),
		SocketGateAddr:  os.Getenv("SOCKETGATE_CONTRACT"),
		ServerPort:      os.Getenv("SERVER_PORT"),
		ContractABI:     `[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"uint256","name":"toChainId","type":"uint256"},{"indexed":false,"internalType":"bytes32","name":"bridgeName","type":"bytes32"},{"indexed":false,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"address","name":"receiver","type":"address"},{"indexed":false,"internalType":"bytes32","name":"metadata","type":"bytes32"}],"name":"SocketBridge","type":"event"}]`,
		TopicHex:        os.Getenv("SOCKET_TOPIC_HEX"),
		CurrencyConfigs: map[string]CurrencyConfig{
			"ETH":     {Factor: 18, Currency: "ETH"},
			"USDT":    {Factor: 16, Currency: "USDT"},
			"DAI":     {Factor: 18, Currency: "DAI"},
			"BTC":     {Factor: 18, Currency: "BTC"},
			"DEFAULT": {Factor: 1, Currency: "WEI"},
		},
	}
}

// GetDefaultCurrency returns default currency `WEI` details
func (c Config) GetDefaultCurrency() CurrencyConfig {
	return c.CurrencyConfigs["DEFAULT"]
}

// GetCurrencyDetails returns the Factor and Currency based on the provided currency
//
// Input currency case insensitive, always converts to uppercase
//
//	currency defaults to `WEI`
func (c Config) GetCurrencyDetails(currency string) CurrencyConfig {
	currConf := c.CurrencyConfigs[strings.ToUpper(currency)]

	// Incase currency is empty or unsupported, it will return default currency
	if currConf.Currency == "" {
		return c.GetDefaultCurrency()
	}

	return currConf
}
