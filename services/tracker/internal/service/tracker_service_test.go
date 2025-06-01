package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sloweyyy/GreenLedger/services/tracker/internal/models"
)

func TestEcoActivityModel_Creation(t *testing.T) {
	activity := &models.EcoActivity{
		ID:             uuid.New(),
		UserID:         "test-user-123",
		ActivityTypeID: uuid.New(),
		Description:    "Bike ride to work",
		Duration:       30,
		Distance:       5.5,
		CreditsEarned:  2.5,
		Source:         models.SourceManual,
	}

	if activity.UserID != "test-user-123" {
		t.Errorf("Expected UserID 'test-user-123', got %s", activity.UserID)
	}

	if activity.Description != "Bike ride to work" {
		t.Errorf("Expected Description 'Bike ride to work', got %s", activity.Description)
	}

	if activity.Duration != 30 {
		t.Errorf("Expected Duration 30, got %d", activity.Duration)
	}

	if activity.Distance != 5.5 {
		t.Errorf("Expected Distance 5.5, got %f", activity.Distance)
	}

	if activity.CreditsEarned != 2.5 {
		t.Errorf("Expected CreditsEarned 2.5, got %f", activity.CreditsEarned)
	}

	if activity.Source != models.SourceManual {
		t.Errorf("Expected Source %s, got %s", models.SourceManual, activity.Source)
	}
}

func TestActivityTypeModel_Creation(t *testing.T) {
	activityType := &models.ActivityType{
		ID:                   uuid.New(),
		Name:                 "Biking",
		Category:             models.CategoryTransport,
		Description:          "Cycling for transportation",
		BaseCreditsPerUnit:   0.5,
		Unit:                 "km",
		IsActive:             true,
		RequiresVerification: false,
	}

	if activityType.Name != "Biking" {
		t.Errorf("Expected Name 'Biking', got %s", activityType.Name)
	}

	if activityType.Category != models.CategoryTransport {
		t.Errorf("Expected Category %s, got %s", models.CategoryTransport, activityType.Category)
	}

	if activityType.BaseCreditsPerUnit != 0.5 {
		t.Errorf("Expected BaseCreditsPerUnit 0.5, got %f", activityType.BaseCreditsPerUnit)
	}

	if !activityType.IsActive {
		t.Error("Expected activity type to be active")
	}
}

func TestChallengeModel_Creation(t *testing.T) {
	challenge := &models.ActivityChallenge{
		ID:            uuid.New(),
		Name:          "30-Day Bike Challenge",
		Description:   "Bike to work for 30 days",
		StartDate:     time.Now(),
		EndDate:       time.Now().AddDate(0, 1, 0),
		TargetValue:   30,
		TargetUnit:    "days",
		RewardCredits: 50.0,
		IsActive:      true,
	}

	if challenge.Name != "30-Day Bike Challenge" {
		t.Errorf("Expected Name '30-Day Bike Challenge', got %s", challenge.Name)
	}

	if challenge.TargetValue != 30 {
		t.Errorf("Expected TargetValue 30, got %f", challenge.TargetValue)
	}

	if challenge.RewardCredits != 50.0 {
		t.Errorf("Expected RewardCredits 50.0, got %f", challenge.RewardCredits)
	}

	if !challenge.IsActive {
		t.Error("Expected challenge to be active")
	}
}

func TestIoTDeviceModel_Creation(t *testing.T) {
	device := &models.IoTDevice{
		ID:       uuid.New(),
		UserID:   "test-user-123",
		DeviceID: "bike-sensor-001",
		Name:     "Bike Sensor",
		Type:     models.DeviceTypeBikeSensor,
		IsActive: true,
		APIKey:   "api-key-123",
	}

	if device.UserID != "test-user-123" {
		t.Errorf("Expected UserID 'test-user-123', got %s", device.UserID)
	}

	if device.DeviceID != "bike-sensor-001" {
		t.Errorf("Expected DeviceID 'bike-sensor-001', got %s", device.DeviceID)
	}

	if device.Type != models.DeviceTypeBikeSensor {
		t.Errorf("Expected Type %s, got %s", models.DeviceTypeBikeSensor, device.Type)
	}

	if !device.IsActive {
		t.Error("Expected device to be active")
	}
}
