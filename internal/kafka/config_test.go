package kafka

import (
	"testing"
	"time"
)

func TestLoadConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv(envKafkaBrokers, "")
	t.Setenv(envKafkaTopic, "")
	t.Setenv(envKafkaClientID, "")
	t.Setenv(envKafkaWriteTimeout, "")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv failed: %v", err)
	}

	if len(cfg.Brokers) != 1 || cfg.Brokers[0] != "localhost:29092" {
		t.Fatalf("unexpected brokers: %#v", cfg.Brokers)
	}
	if cfg.Topic != defaultTopic {
		t.Fatalf("expected topic %q, got %q", defaultTopic, cfg.Topic)
	}
	if cfg.ClientID != defaultClientID {
		t.Fatalf("expected client id %q, got %q", defaultClientID, cfg.ClientID)
	}
	if cfg.WriteTimeout != defaultWriteTimeout {
		t.Fatalf("expected timeout %s, got %s", defaultWriteTimeout, cfg.WriteTimeout)
	}
}

func TestLoadConfigFromEnv_Overrides(t *testing.T) {
	t.Setenv(envKafkaBrokers, "localhost:29092, kafka:9092 ")
	t.Setenv(envKafkaTopic, "custom-topic")
	t.Setenv(envKafkaClientID, "custom-client")
	t.Setenv(envKafkaWriteTimeout, "2500")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv failed: %v", err)
	}

	if len(cfg.Brokers) != 2 || cfg.Brokers[0] != "localhost:29092" || cfg.Brokers[1] != "kafka:9092" {
		t.Fatalf("unexpected brokers: %#v", cfg.Brokers)
	}
	if cfg.Topic != "custom-topic" {
		t.Fatalf("unexpected topic: %q", cfg.Topic)
	}
	if cfg.ClientID != "custom-client" {
		t.Fatalf("unexpected client id: %q", cfg.ClientID)
	}
	if cfg.WriteTimeout != 2500*time.Millisecond {
		t.Fatalf("unexpected timeout: %s", cfg.WriteTimeout)
	}
}

func TestLoadConfigFromEnv_InvalidTimeout(t *testing.T) {
	t.Setenv(envKafkaWriteTimeout, "not-a-number")

	if _, err := LoadConfigFromEnv(); err == nil {
		t.Fatal("expected invalid timeout error")
	}
}
