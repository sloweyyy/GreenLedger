package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sloweyyy/GreenLedger/shared/logger"
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
		Topic:        "greenledger-events",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	return &KafkaEventPublisher{
		writer: writer,
		logger: logger,
	}
}

// PublishCreditEarned publishes a credit earned event
func (p *KafkaEventPublisher) PublishCreditEarned(ctx context.Context, event *CreditEarnedEvent) error {
	// Add event metadata
	eventWithMetadata := struct {
		*CreditEarnedEvent
		EventType string    `json:"event_type"`
		EventID   string    `json:"event_id"`
		Source    string    `json:"source"`
		Version   string    `json:"version"`
		Timestamp time.Time `json:"timestamp"`
	}{
		CreditEarnedEvent: event,
		EventType:         "credit_earned",
		EventID:           fmt.Sprintf("credit_%s_%d", event.UserID, time.Now().UnixNano()),
		Source:            "tracker-service",
		Version:           "1.0",
		Timestamp:         event.Timestamp,
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
			{Key: "event-type", Value: []byte("credit_earned")},
			{Key: "user-id", Value: []byte(event.UserID)},
			{Key: "source", Value: []byte("tracker-service")},
		},
	}

	// Publish message
	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		p.logger.LogError(ctx, "failed to publish credit earned event", err,
			logger.String("user_id", event.UserID),
			logger.String("activity_id", event.ActivityID))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.LogInfo(ctx, "credit earned event published",
		logger.String("user_id", event.UserID),
		logger.String("activity_id", event.ActivityID),
		logger.Float64("credits", event.CreditsEarned))

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

// PublishCreditEarned publishes a credit earned event (mock)
func (p *MockEventPublisher) PublishCreditEarned(ctx context.Context, event *CreditEarnedEvent) error {
	p.Events = append(p.Events, event)
	p.logger.LogInfo(ctx, "mock: credit earned event published",
		logger.String("user_id", event.UserID),
		logger.String("activity_id", event.ActivityID))
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
