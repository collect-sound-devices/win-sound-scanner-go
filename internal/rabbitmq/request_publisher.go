package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RequestPublisher manages RabbitMQ connection, topology, and message publishing.
type RequestPublisher struct {
	cfg    Config
	logger *slog.Logger

	mu       sync.Mutex
	conn     *amqp.Connection
	ch       *amqp.Channel
	confirms <-chan amqp.Confirmation
}

func NewRequestPublisher(ctx context.Context, cfg Config, logger *slog.Logger) (*RequestPublisher, error) {
	if ctx == nil {
		panic("nil context")
	}
	if logger == nil {
		panic("nil logger")
	}

	cfg = cfg.withDefaults()

	p := &RequestPublisher{
		cfg:    cfg,
		logger: logger,
	}

	if err := p.connectWithRetryLocked(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *RequestPublisher) Publish(ctx context.Context, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ctx == nil {
		panic("nil context")
	}

	if p.ch == nil {
		if err := p.connectWithRetryLocked(ctx); err != nil {
			return err
		}
	}

	if err := p.publishLocked(ctx, body); err == nil {
		return nil
	} else {
		p.logger.Warn("RabbitMQ publish failed, reconnecting once", "err", err)
		if recErr := p.connectWithRetryLocked(ctx); recErr != nil {
			return fmt.Errorf("rabbitmq publish failed: %w (reconnect failed: %v)", err, recErr)
		}
		if retryErr := p.publishLocked(ctx, body); retryErr != nil {
			return fmt.Errorf("rabbitmq publish failed after reconnect: %w", retryErr)
		}
	}

	return nil
}

func (p *RequestPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closeLocked()
}

func (p *RequestPublisher) publishLocked(ctx context.Context, body []byte) error {
	if p.ch == nil {
		return errors.New("rabbitmq channel is not initialized")
	}

	err := p.ch.PublishWithContext(
		ctx,
		p.cfg.ExchangeName,
		p.cfg.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish call failed: %w", err)
	}

	confirmTimeout := p.cfg.PublishConfirmTimeout
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return context.DeadlineExceeded
		}
		if remaining < confirmTimeout {
			confirmTimeout = remaining
		}
	}

	timer := time.NewTimer(confirmTimeout)
	defer timer.Stop()

	select {
	case c, ok := <-p.confirms:
		if !ok {
			return errors.New("rabbitmq confirms channel is closed")
		}
		if !c.Ack {
			return fmt.Errorf("message NOT ACKed (deliveryTag=%d)", c.DeliveryTag)
		}
		p.logger.Info("RabbitMQ message ACKed", "routingKey", p.cfg.RoutingKey, "deliveryTag", c.DeliveryTag)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return fmt.Errorf("timed out waiting for publish confirmation after %s", confirmTimeout)
	}
}

func (p *RequestPublisher) connectWithRetryLocked(ctx context.Context) error {
	var lastErr error
	delay := p.cfg.InitialReconnectDelay

	for attempt := 1; attempt <= p.cfg.MaxReconnectionAttempts; attempt++ {
		if err := p.connectOnceLocked(); err == nil {
			p.logger.Info("RabbitMQ producer initialized", "attempt", attempt)
			return nil
		} else {
			lastErr = err
			if attempt == p.cfg.MaxReconnectionAttempts {
				break
			}
			p.logger.Warn(
				"RabbitMQ init attempt failed; retrying",
				"attempt",
				attempt,
				"maxAttempts",
				p.cfg.MaxReconnectionAttempts,
				"retryDelay",
				delay,
				"err",
				err,
			)

			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return ctx.Err()
			case <-timer.C:
			}

			delay = minDuration(delay*2, p.cfg.MaxReconnectDelay)
		}
	}

	return fmt.Errorf("rabbitmq initialization failed after %d attempts: %w", p.cfg.MaxReconnectionAttempts, lastErr)
}

func (p *RequestPublisher) connectOnceLocked() error {
	_ = p.closeLocked()

	conn, err := amqp.DialConfig(
		amqp.URI{
			Scheme:   "amqp",
			Host:     p.cfg.Host,
			Port:     p.cfg.Port,
			Username: p.cfg.User,
			Password: p.cfg.Password,
			Vhost:    p.cfg.VHost,
		}.String(),
		amqp.Config{Heartbeat: p.cfg.ConnectionThreshold},
	)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("channel open failed: %w", err)
	}

	if err := ch.ExchangeDeclare(
		p.cfg.ExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("exchange declare failed: %w", err)
	}

	q, err := ch.QueueDeclare(
		p.cfg.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("queue declare failed: %w", err)
	}

	if err := ch.QueueBind(
		q.Name,
		p.cfg.RoutingKey,
		p.cfg.ExchangeName,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("queue bind failed: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("confirm mode failed: %w", err)
	}

	p.conn = conn
	p.ch = ch
	p.confirms = ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	return nil
}

func (p *RequestPublisher) closeLocked() error {
	var err error

	if p.ch != nil {
		err = errors.Join(err, p.ch.Close())
		p.ch = nil
	}
	if p.conn != nil {
		err = errors.Join(err, p.conn.Close())
		p.conn = nil
	}
	p.confirms = nil

	return err
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
