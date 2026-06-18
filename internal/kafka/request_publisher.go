package kafka

import (
	"context"
	"fmt"
	"log/slog"

	kafkago "github.com/segmentio/kafka-go"
)

type RequestPublisher struct {
	cfg    Config
	logger *slog.Logger
	writer *kafkago.Writer
}

func NewRequestPublisher(cfg Config, logger *slog.Logger) (*RequestPublisher, error) {
	if logger == nil {
		panic("nil logger")
	}

	cfg = cfg.withDefaults()
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are empty")
	}

	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafkago.Hash{},
		BatchSize:    1,
		RequiredAcks: kafkago.RequireAll,
		WriteTimeout: cfg.WriteTimeout,
		Async:        false,
		Transport: &kafkago.Transport{
			ClientID: cfg.ClientID,
		},
	}

	logger.Info("Kafka producer initialized", "brokers", cfg.Brokers, "topic", cfg.Topic, "clientId", cfg.ClientID)
	return &RequestPublisher{
		cfg:    cfg,
		logger: logger,
		writer: writer,
	}, nil
}

func (p *RequestPublisher) Publish(ctx context.Context, key []byte, body []byte) error {
	if ctx == nil {
		panic("nil context")
	}

	if err := p.writer.WriteMessages(ctx, kafkago.Message{Key: key, Value: body}); err != nil {
		return fmt.Errorf("kafka write failed: %w", err)
	}

	p.logger.Info("Kafka message written", "topic", p.cfg.Topic, "key", string(key))
	return nil
}

func (p *RequestPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
