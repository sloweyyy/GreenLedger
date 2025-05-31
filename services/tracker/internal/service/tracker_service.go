package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greenledger/services/tracker/internal/models"
	"github.com/greenledger/services/tracker/internal/repository"
	"github.com/greenledger/shared/logger"
)

// TrackerService handles eco-activity tracking operations
type TrackerService struct {
	activityRepo     *repository.ActivityRepository
	activityTypeRepo *repository.ActivityTypeRepository
	creditRuleRepo   *repository.CreditRuleRepository
	eventPublisher   EventPublisher
	logger           *logger.Logger
}

// NewTrackerService creates a new tracker service
func NewTrackerService(
	activityRepo *repository.ActivityRepository,
	activityTypeRepo *repository.ActivityTypeRepository,
	creditRuleRepo *repository.CreditRuleRepository,
	eventPublisher EventPublisher,
	logger *logger.Logger,
) *TrackerService {
	return &TrackerService{
		activityRepo:     activityRepo,
		activityTypeRepo: activityTypeRepo,
		creditRuleRepo:   creditRuleRepo,
		eventPublisher:   eventPublisher,
		logger:           logger,
	}
}

// LogActivityRequest represents a request to log an eco-activity
type LogActivityRequest struct {
	UserID         string                 `json:"user_id" binding:"required"`
	ActivityType   string                 `json:"activity_type" binding:"required"`
	Description    string                 `json:"description" binding:"required"`
	Duration       int                    `json:"duration"` // in minutes
	Distance       float64                `json:"distance"` // in kilometers
	Quantity       float64                `json:"quantity"`
	Unit           string                 `json:"unit"`
	Location       string                 `json:"location"`
	Source         string                 `json:"source"`
	SourceData     map[string]interface{} `json:"source_data"`
}

// ActivityResponse represents an activity in API responses
type ActivityResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        string    `json:"user_id"`
	ActivityType  string    `json:"activity_type"`
	Description   string    `json:"description"`
	Duration      int       `json:"duration"`
	Distance      float64   `json:"distance"`
	Quantity      float64   `json:"quantity"`
	Unit          string    `json:"unit"`
	Location      string    `json:"location"`
	CreditsEarned float64   `json:"credits_earned"`
	IsVerified    bool      `json:"is_verified"`
	Source        string    `json:"source"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreditEarnedEvent represents an event when credits are earned
type CreditEarnedEvent struct {
	UserID        string    `json:"user_id"`
	ActivityID    string    `json:"activity_id"`
	ActivityType  string    `json:"activity_type"`
	CreditsEarned float64   `json:"credits_earned"`
	Description   string    `json:"description"`
	Timestamp     time.Time `json:"timestamp"`
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	PublishCreditEarned(ctx context.Context, event *CreditEarnedEvent) error
}

// LogActivity logs a new eco-friendly activity
func (s *TrackerService) LogActivity(ctx context.Context, req *LogActivityRequest) (*ActivityResponse, error) {
	s.logger.LogInfo(ctx, "logging eco-activity",
		logger.String("user_id", req.UserID),
		logger.String("activity_type", req.ActivityType))

	// Get activity type
	activityType, err := s.activityTypeRepo.GetByName(ctx, req.ActivityType)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity type: %w", err)
	}

	if !activityType.IsActive {
		return nil, fmt.Errorf("activity type is not active")
	}

	// Calculate credits earned
	creditsEarned, err := s.calculateCredits(ctx, activityType, req)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate credits: %w", err)
	}

	// Convert source data to JSON
	sourceDataJSON := ""
	if req.SourceData != nil {
		sourceDataBytes, err := json.Marshal(req.SourceData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal source data: %w", err)
		}
		sourceDataJSON = string(sourceDataBytes)
	}

	// Create activity
	activity := &models.EcoActivity{
		UserID:         req.UserID,
		ActivityTypeID: activityType.ID,
		Description:    req.Description,
		Duration:       req.Duration,
		Distance:       req.Distance,
		Quantity:       req.Quantity,
		Unit:           req.Unit,
		Location:       req.Location,
		CreditsEarned:  creditsEarned,
		IsVerified:     !activityType.RequiresVerification,
		Source:         req.Source,
		SourceData:     sourceDataJSON,
	}

	if req.Source == "" {
		activity.Source = models.SourceManual
	}

	// Save activity
	if err := s.activityRepo.Create(ctx, activity); err != nil {
		return nil, fmt.Errorf("failed to save activity: %w", err)
	}

	// Publish credit earned event if verified
	if activity.IsVerified && creditsEarned > 0 {
		event := &CreditEarnedEvent{
			UserID:        activity.UserID,
			ActivityID:    activity.ID.String(),
			ActivityType:  activityType.Name,
			CreditsEarned: creditsEarned,
			Description:   activity.Description,
			Timestamp:     time.Now().UTC(),
		}

		if err := s.eventPublisher.PublishCreditEarned(ctx, event); err != nil {
			s.logger.LogError(ctx, "failed to publish credit earned event", err,
				logger.String("activity_id", activity.ID.String()))
			// Don't fail the request, just log the error
		}
	}

	s.logger.LogInfo(ctx, "eco-activity logged successfully",
		logger.String("activity_id", activity.ID.String()),
		logger.String("user_id", req.UserID),
		logger.Float64("credits_earned", creditsEarned))

	return s.activityToResponse(activity, activityType), nil
}

// GetUserActivities retrieves activities for a user
func (s *TrackerService) GetUserActivities(ctx context.Context, userID string, limit, offset int) ([]*ActivityResponse, int64, error) {
	activities, total, err := s.activityRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user activities: %w", err)
	}

	responses := make([]*ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = s.activityToResponse(activity, &activity.ActivityType)
	}

	return responses, total, nil
}

// GetActivityByID retrieves a specific activity
func (s *TrackerService) GetActivityByID(ctx context.Context, id uuid.UUID) (*ActivityResponse, error) {
	activity, err := s.activityRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return s.activityToResponse(activity, &activity.ActivityType), nil
}

// VerifyActivity verifies an activity (admin/moderator operation)
func (s *TrackerService) VerifyActivity(ctx context.Context, activityID uuid.UUID, verifiedBy string) error {
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return fmt.Errorf("failed to get activity: %w", err)
	}

	if activity.IsVerified {
		return fmt.Errorf("activity is already verified")
	}

	now := time.Now().UTC()
	activity.IsVerified = true
	activity.VerifiedAt = &now
	activity.VerifiedBy = verifiedBy

	if err := s.activityRepo.Update(ctx, activity); err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}

	// Publish credit earned event now that it's verified
	if activity.CreditsEarned > 0 {
		event := &CreditEarnedEvent{
			UserID:        activity.UserID,
			ActivityID:    activity.ID.String(),
			ActivityType:  activity.ActivityType.Name,
			CreditsEarned: activity.CreditsEarned,
			Description:   activity.Description,
			Timestamp:     now,
		}

		if err := s.eventPublisher.PublishCreditEarned(ctx, event); err != nil {
			s.logger.LogError(ctx, "failed to publish credit earned event", err,
				logger.String("activity_id", activity.ID.String()))
		}
	}

	s.logger.LogInfo(ctx, "activity verified",
		logger.String("activity_id", activityID.String()),
		logger.String("verified_by", verifiedBy))

	return nil
}

// GetUserStats retrieves activity statistics for a user
func (s *TrackerService) GetUserStats(ctx context.Context, userID string, startDate, endDate time.Time) (*UserActivityStats, error) {
	stats, err := s.activityRepo.GetUserStats(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}

// calculateCredits calculates credits earned for an activity
func (s *TrackerService) calculateCredits(ctx context.Context, activityType *models.ActivityType, req *LogActivityRequest) (float64, error) {
	// Get applicable credit rules
	rules, err := s.creditRuleRepo.GetActiveRulesByActivityType(ctx, activityType.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get credit rules: %w", err)
	}

	// If no specific rules, use base credits
	if len(rules) == 0 {
		return s.calculateBaseCredits(activityType, req), nil
	}

	// Find the best matching rule
	var bestRule *models.CreditRule
	var value float64

	// Determine the value to use for rule matching
	switch activityType.Unit {
	case "minutes":
		value = float64(req.Duration)
	case "km":
		value = req.Distance
	case "units":
		value = req.Quantity
	default:
		value = req.Quantity
	}

	for _, rule := range rules {
		if value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue) {
			if bestRule == nil || rule.CreditsPerUnit > bestRule.CreditsPerUnit {
				bestRule = rule
			}
		}
	}

	if bestRule != nil {
		return value * bestRule.CreditsPerUnit * bestRule.Multiplier, nil
	}

	// Fallback to base credits
	return s.calculateBaseCredits(activityType, req), nil
}

// calculateBaseCredits calculates base credits without rules
func (s *TrackerService) calculateBaseCredits(activityType *models.ActivityType, req *LogActivityRequest) float64 {
	switch activityType.Unit {
	case "minutes":
		return float64(req.Duration) * activityType.BaseCreditsPerUnit
	case "km":
		return req.Distance * activityType.BaseCreditsPerUnit
	case "units":
		return req.Quantity * activityType.BaseCreditsPerUnit
	default:
		return req.Quantity * activityType.BaseCreditsPerUnit
	}
}

// activityToResponse converts an activity model to response format
func (s *TrackerService) activityToResponse(activity *models.EcoActivity, activityType *models.ActivityType) *ActivityResponse {
	return &ActivityResponse{
		ID:            activity.ID,
		UserID:        activity.UserID,
		ActivityType:  activityType.Name,
		Description:   activity.Description,
		Duration:      activity.Duration,
		Distance:      activity.Distance,
		Quantity:      activity.Quantity,
		Unit:          activity.Unit,
		Location:      activity.Location,
		CreditsEarned: activity.CreditsEarned,
		IsVerified:    activity.IsVerified,
		Source:        activity.Source,
		CreatedAt:     activity.CreatedAt,
	}
}

// UserActivityStats represents activity statistics for a user
type UserActivityStats struct {
	UserID            string    `json:"user_id"`
	TotalActivities   int64     `json:"total_activities"`
	TotalCreditsEarned float64  `json:"total_credits_earned"`
	TotalDuration     int       `json:"total_duration"`
	TotalDistance     float64   `json:"total_distance"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
}
