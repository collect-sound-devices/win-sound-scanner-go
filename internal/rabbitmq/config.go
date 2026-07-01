package rabbitmq

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHost                    = "localhost"
	defaultPort                    = 5672
	defaultVHost                   = "/"
	defaultUser                    = "guest"
	defaultPassword                = "guest"
	defaultExchangeName            = "sdr_exchange"
	defaultQueueName               = "sdr_queue"
	defaultRoutingKey              = "sdr_bind"
	defaultConnectionThreshold     = 20 * time.Second
	defaultMaxReconnectionAttempts = 8
	defaultInitialReconnectDelay   = 2 * time.Second
	defaultMaxReconnectDelay       = 30 * time.Second
	defaultPublishConfirmTimeout   = 10 * time.Second
)

// Config defines RabbitMQ connection, topology, and retry settings.
type Config struct {
	Host                    string
	Port                    int
	VHost                   string
	User                    string
	Password                string
	ExchangeName            string
	QueueName               string
	RoutingKey              string
	ConnectionThreshold     time.Duration
	MaxReconnectionAttempts int
	InitialReconnectDelay   time.Duration
	MaxReconnectDelay       time.Duration
	PublishConfirmTimeout   time.Duration
}

func DefaultConfig() Config {
	return Config{
		Host:                    defaultHost,
		Port:                    defaultPort,
		VHost:                   defaultVHost,
		User:                    defaultUser,
		Password:                defaultPassword,
		ExchangeName:            defaultExchangeName,
		QueueName:               defaultQueueName,
		RoutingKey:              defaultRoutingKey,
		ConnectionThreshold:     defaultConnectionThreshold,
		MaxReconnectionAttempts: defaultMaxReconnectionAttempts,
		InitialReconnectDelay:   defaultInitialReconnectDelay,
		MaxReconnectDelay:       defaultMaxReconnectDelay,
		PublishConfirmTimeout:   defaultPublishConfirmTimeout,
	}
}

func (c Config) withDefaults() Config {
	d := DefaultConfig()

	if host, port, ok := splitHostPort(c.Host); ok {
		c.Host = host
		if c.Port <= 0 || c.Port == d.Port {
			c.Port = port
		}
	}

	if strings.TrimSpace(c.Host) == "" {
		c.Host = d.Host
	}
	if c.Port <= 0 {
		c.Port = d.Port
	}
	if c.VHost == "" {
		c.VHost = d.VHost
	}
	if c.User == "" {
		c.User = d.User
	}
	if c.Password == "" {
		c.Password = d.Password
	}
	if c.ExchangeName == "" {
		c.ExchangeName = d.ExchangeName
	}
	if c.QueueName == "" {
		c.QueueName = d.QueueName
	}
	if c.RoutingKey == "" {
		c.RoutingKey = d.RoutingKey
	}
	if c.ConnectionThreshold <= 0 {
		c.ConnectionThreshold = d.ConnectionThreshold
	}
	if c.MaxReconnectionAttempts <= 0 {
		c.MaxReconnectionAttempts = d.MaxReconnectionAttempts
	}
	if c.InitialReconnectDelay <= 0 {
		c.InitialReconnectDelay = d.InitialReconnectDelay
	}
	if c.MaxReconnectDelay <= 0 {
		c.MaxReconnectDelay = d.MaxReconnectDelay
	}
	if c.PublishConfirmTimeout <= 0 {
		c.PublishConfirmTimeout = d.PublishConfirmTimeout
	}
	if c.MaxReconnectDelay < c.InitialReconnectDelay {
		c.MaxReconnectDelay = c.InitialReconnectDelay
	}

	return c
}

// LoadConfigFromEnv loads rabbit configuration from environment variables.
// Empty values are replaced by defaults.
func LoadConfigFromEnv() (Config, error) {
	cfg := DefaultConfig()

	cfg.Host = trimmedEnvOrDefault("WIN_SOUND_RABBITMQ_HOST", cfg.Host)
	cfg.VHost = envOrDefault("WIN_SOUND_RABBITMQ_VHOST", cfg.VHost)
	cfg.User = envOrDefault("WIN_SOUND_RABBITMQ_USER", cfg.User)
	cfg.Password = envOrDefault("WIN_SOUND_RABBITMQ_PASSWORD", cfg.Password)
	cfg.ExchangeName = envOrDefault("WIN_SOUND_RABBITMQ_EXCHANGE", cfg.ExchangeName)
	cfg.QueueName = envOrDefault("WIN_SOUND_RABBITMQ_QUEUE", cfg.QueueName)
	cfg.RoutingKey = envOrDefault("WIN_SOUND_RABBITMQ_ROUTING_KEY", cfg.RoutingKey)

	port, err := intEnvOrDefault("WIN_SOUND_RABBITMQ_PORT", cfg.Port)
	if err != nil {
		return Config{}, err
	}
	cfg.Port = port

	connectionThresholdSeconds, err := intEnvOrDefault("WIN_SOUND_RABBITMQ_CONNECTION_THRESHOLD_SEC", int(cfg.ConnectionThreshold/time.Second))
	if err != nil {
		return Config{}, err
	}
	cfg.ConnectionThreshold = time.Duration(connectionThresholdSeconds) * time.Second

	maxReconnectAttempts, err := nonNegativeIntEnvOrDefault("WIN_SOUND_RABBITMQ_MAX_RECONNECT_ATTEMPTS", cfg.MaxReconnectionAttempts)
	if err != nil {
		return Config{}, err
	}
	cfg.MaxReconnectionAttempts = maxReconnectAttempts

	initialReconnectDelayMillis, err := intEnvOrDefault("WIN_SOUND_RABBITMQ_INITIAL_RECONNECT_DELAY_MS", int(cfg.InitialReconnectDelay/time.Millisecond))
	if err != nil {
		return Config{}, err
	}
	cfg.InitialReconnectDelay = time.Duration(initialReconnectDelayMillis) * time.Millisecond

	maxReconnectDelayMillis, err := nonNegativeIntEnvOrDefault("WIN_SOUND_RABBITMQ_MAX_RECONNECT_DELAY_MS", int(cfg.MaxReconnectDelay/time.Millisecond))
	if err != nil {
		return Config{}, err
	}
	cfg.MaxReconnectDelay = time.Duration(maxReconnectDelayMillis) * time.Millisecond

	publishConfirmTimeoutMillis, err := nonNegativeIntEnvOrDefault("WIN_SOUND_RABBITMQ_PUBLISH_CONFIRM_TIMEOUT_MS", int(cfg.PublishConfirmTimeout/time.Millisecond))
	if err != nil {
		return Config{}, err
	}
	cfg.PublishConfirmTimeout = time.Duration(publishConfirmTimeoutMillis) * time.Millisecond

	return cfg.withDefaults(), nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func trimmedEnvOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func intEnvOrDefault(key string, fallback int) (int, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback, nil
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, v, err)
	}

	return n, nil
}

func nonNegativeIntEnvOrDefault(key string, fallback int) (int, error) {
	n, err := intEnvOrDefault(key, fallback)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, fmt.Errorf("%s can not be negative %q", key, strings.TrimSpace(os.Getenv(key)))
	}
	return n, nil
}

func splitHostPort(raw string) (string, int, bool) {
	host, portText, err := net.SplitHostPort(strings.TrimSpace(raw))
	if err != nil {
		return "", 0, false
	}

	port, err := strconv.Atoi(portText)
	if err != nil || port <= 0 {
		return "", 0, false
	}
	return host, port, true
}
