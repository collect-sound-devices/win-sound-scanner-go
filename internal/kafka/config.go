package kafka

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBrokers       = "localhost:29092"
	defaultTopic         = "audio-device-events"
	defaultClientID      = "win-sound-scanner"
	defaultWriteTimeout  = 10 * time.Second
	envKafkaBrokers      = "WIN_SOUND_KAFKA_BROKERS"
	envKafkaTopic        = "WIN_SOUND_KAFKA_TOPIC"
	envKafkaClientID     = "WIN_SOUND_KAFKA_CLIENT_ID"
	envKafkaWriteTimeout = "WIN_SOUND_KAFKA_WRITE_TIMEOUT_MS"
)

type Config struct {
	Brokers      []string
	Topic        string
	ClientID     string
	WriteTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{
		Brokers:      splitCSV(defaultBrokers),
		Topic:        defaultTopic,
		ClientID:     defaultClientID,
		WriteTimeout: defaultWriteTimeout,
	}
}

func (c Config) withDefaults() Config {
	d := DefaultConfig()
	if len(c.Brokers) == 0 {
		c.Brokers = d.Brokers
	}
	if strings.TrimSpace(c.Topic) == "" {
		c.Topic = d.Topic
	}
	if strings.TrimSpace(c.ClientID) == "" {
		c.ClientID = d.ClientID
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = d.WriteTimeout
	}
	return c
}

func LoadConfigFromEnv() (Config, error) {
	cfg := DefaultConfig()

	if v := strings.TrimSpace(os.Getenv(envKafkaBrokers)); v != "" {
		cfg.Brokers = splitCSV(v)
	}
	if v := strings.TrimSpace(os.Getenv(envKafkaTopic)); v != "" {
		cfg.Topic = v
	}
	if v := strings.TrimSpace(os.Getenv(envKafkaClientID)); v != "" {
		cfg.ClientID = v
	}
	if v := strings.TrimSpace(os.Getenv(envKafkaWriteTimeout)); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid %s %q: %w", envKafkaWriteTimeout, v, err)
		}
		if n < 0 {
			return Config{}, fmt.Errorf("%s can not be negative %q", envKafkaWriteTimeout, v)
		}
		cfg.WriteTimeout = time.Duration(n) * time.Millisecond
	}

	return cfg.withDefaults(), nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
