package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppEnv                AppEnv `env:"APP_ENV"`
	AppPort               uint16 `env:"APP_PORT"`
	DBHost                string `env:"DB_HOST"`
	DBPort                uint16 `env:"DB_PORT"`
	DBUsername            string `env:"DB_USERNAME"`
	DBPassword            string `env:"DB_PASSWORD"`
	DBName                string `env:"DB_NAME"`
	DBPath                string `env:"DB_PATH"`
	MaxBatchSize          uint   `env:"MAX_BATCH_SIZE"`
	MaxRetries            uint8  `env:"MAX_RETRIES"`
	KafkaPort             uint16 `env:"KAFKA_PORT"`
	NotificationTopicName string `env:"NOTIFICATION_TOPIC_NAME"`
	ConsumerGroupID       string `env:"CONSUMER_GROUP_ID"`
	SenderHandlePeriodMs  int    `env:"SENDER_HANDLE_PERIOD_MS"`
	Timeout               int    `env:"TIMEOUT"`
	Gmail                 string `env:"GMAIL"`
	GmailAppPassword      string `env:"GMAIL_APP_PASSWORD"`
}

type AppEnv string

const (
	Local AppEnv = "local"
	Dev   AppEnv = "dev"
	Prod  AppEnv = "prod"
)

var (
	once sync.Once
	Cfg  *Config
)

func MustLoad() *Config {
	once.Do(func() {
		Cfg = &Config{}
		if err := cleanenv.ReadEnv(Cfg); err != nil {
			log.Fatalf("Cannot read .env file: %s", err)
		}
		fmt.Printf("APP_PORT: %d", Cfg.AppPort)
	})
	return Cfg
}

func (cfg *Config) GetDBURL() string {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBName,
	)
	return dbUrl
}
