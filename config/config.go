package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

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
	}
}
