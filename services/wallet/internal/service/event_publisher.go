package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/greenledger/shared/logger"
	"github.com/segmentio/kafka-go"
)

// KafkaEventPublisher implements EventPublisher using Kafka
type KafkaEventPublisher struct {
	writer *kafka.Writer
	logger *logger.Logger
}

// NewKafkaEventPublisher creates a new Kafka event publisher
func NewKafkaEventPublisher(brokers []string, logger *logger.Logger) *KafkaEventPublisher {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        "greenledger-wallet-events",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	return &KafkaEventPublisher{
		writer: writer,
		logger: logger,
	}
}

// PublishBalanceUpdated publishes a balance updated event
func (p *KafkaEventPublisher) PublishBalanceUpdated(ctx context.Context, event *BalanceUpdatedEvent) error {
	// Add event metadata
	eventWithMetadata := struct {
		*BalanceUpdatedEvent
		EventType string    `json:"event_type"`
		EventID   string    `json:"event_id"`
		Source    string    `json:"source"`
		Version   string    `json:"version"`
		Timestamp time.Time `json:"timestamp"`
	}{
		BalanceUpdatedEvent: event,
		EventType:           "balance_updated",
		EventID:             fmt.Sprintf("balance_%s_%d", event.UserID, time.Now().UnixNano()),
		Source:              "wallet-service",
		Version:             "1.0",
		Timestamp:           event.Timestamp,
	}

	// Serialize event
	eventData, err := json.Marshal(eventWithMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := kafka.Message{
		Key:   []byte(event.UserID),
		Value: eventData,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte("balance_updated")},
			{Key: "user-id", Value: []byte(event.UserID)},
			{Key: "source", Value: []byte("wallet-service")},
		},
	}

	// Publish message
	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		p.logger.LogError(ctx, "failed to publish balance updated event", err,
			logger.String("user_id", event.UserID),
			logger.String("transaction_id", event.TransactionID))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.LogInfo(ctx, "balance updated event published",
		logger.String("user_id", event.UserID),
		logger.String("transaction_id", event.TransactionID))

	return nil
}

// PublishTransferCompleted publishes a transfer completed event
func (p *KafkaEventPublisher) PublishTransferCompleted(ctx context.Context, event *TransferCompletedEvent) error {
	// Add event metadata
	eventWithMetadata := struct {
		*TransferCompletedEvent
		EventType string    `json:"event_type"`
		EventID   string    `json:"event_id"`
		Source    string    `json:"source"`
		Version   string    `json:"version"`
		Timestamp time.Time `json:"timestamp"`
	}{
		TransferCompletedEvent: event,
		EventType:              "transfer_completed",
		EventID:                fmt.Sprintf("transfer_%s_%d", event.TransferID, time.Now().UnixNano()),
		Source:                 "wallet-service",
		Version:                "1.0",
		Timestamp:              event.Timestamp,
	}

	// Serialize event
	eventData, err := json.Marshal(eventWithMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := kafka.Message{
		Key:   []byte(event.TransferID),
		Value: eventData,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte("transfer_completed")},
			{Key: "from-user-id", Value: []byte(event.FromUserID)},
			{Key: "to-user-id", Value: []byte(event.ToUserID)},
			{Key: "source", Value: []byte("wallet-service")},
		},
	}

	// Publish message
	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		p.logger.LogError(ctx, "failed to publish transfer completed event", err,
			logger.String("transfer_id", event.TransferID),
			logger.String("from_user_id", event.FromUserID),
			logger.String("to_user_id", event.ToUserID))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.LogInfo(ctx, "transfer completed event published",
		logger.String("transfer_id", event.TransferID),
		logger.String("from_user_id", event.FromUserID),
		logger.String("to_user_id", event.ToUserID))

	return nil
}

// Close closes the event publisher
func (p *KafkaEventPublisher) Close() error {
	return p.writer.Close()
}

// MockEventPublisher is a mock implementation for testing
type MockEventPublisher struct {
	Events []interface{}
	logger *logger.Logger
}

// NewMockEventPublisher creates a new mock event publisher
func NewMockEventPublisher(logger *logger.Logger) *MockEventPublisher {
	return &MockEventPublisher{
		Events: make([]interface{}, 0),
		logger: logger,
	}
}

// PublishBalanceUpdated publishes a balance updated event (mock)
func (p *MockEventPublisher) PublishBalanceUpdated(ctx context.Context, event *BalanceUpdatedEvent) error {
	p.Events = append(p.Events, event)
	p.logger.LogInfo(ctx, "mock: balance updated event published",
		logger.String("user_id", event.UserID),
		logger.String("transaction_id", event.TransactionID))
	return nil
}

// PublishTransferCompleted publishes a transfer completed event (mock)
func (p *MockEventPublisher) PublishTransferCompleted(ctx context.Context, event *TransferCompletedEvent) error {
	p.Events = append(p.Events, event)
	p.logger.LogInfo(ctx, "mock: transfer completed event published",
		logger.String("transfer_id", event.TransferID),
		logger.String("from_user_id", event.FromUserID),
		logger.String("to_user_id", event.ToUserID))
	return nil
}

// GetEvents returns all published events
func (p *MockEventPublisher) GetEvents() []interface{} {
	return p.Events
}

// Clear clears all events
func (p *MockEventPublisher) Clear() {
	p.Events = make([]interface{}, 0)
}

// EventConsumer handles consuming wallet events
type EventConsumer struct {
	reader *kafka.Reader
	logger *logger.Logger
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(brokers []string, groupID string, logger *logger.Logger) *EventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "greenledger-events",
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &EventConsumer{
		reader: reader,
		logger: logger,
	}
}

// ConsumeEvents consumes events from Kafka
func (c *EventConsumer) ConsumeEvents(ctx context.Context, handler func(ctx context.Context, event interface{}) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.LogError(ctx, "failed to read message", err)
				continue
			}

			// Parse event type from headers
			var eventType string
			for _, header := range message.Headers {
				if header.Key == "event-type" {
					eventType = string(header.Value)
					break
				}
			}

			// Handle different event types
			switch eventType {
			case "credit_earned":
				var event CreditEarnedEvent
				if err := json.Unmarshal(message.Value, &event); err != nil {
					c.logger.LogError(ctx, "failed to unmarshal credit earned event", err)
					continue
				}
				
				// Process credit earned event - credit user's wallet
				if err := handler(ctx, &event); err != nil {
					c.logger.LogError(ctx, "failed to handle credit earned event", err)
				}

			default:
				c.logger.LogWarn(ctx, "unknown event type",
					logger.String("event_type", eventType))
			}
		}
	}
}

// Close closes the event consumer
func (c *EventConsumer) Close() error {
	return c.reader.Close()
}

// CreditEarnedEvent represents a credit earned event from tracker service
type CreditEarnedEvent struct {
	UserID        string  `json:"user_id"`
	ActivityID    string  `json:"activity_id"`
	ActivityType  string  `json:"activity_type"`
	CreditsEarned float64 `json:"credits_earned"`
	Description   string  `json:"description"`
	Timestamp     time.Time `json:"timestamp"`
}
