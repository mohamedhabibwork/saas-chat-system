package services

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedhabibwork/saas-chat-system/internal/database"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
)

// LocationService handles location-related operations
type LocationService struct {
	db *database.DB
}

// NewLocationService creates a new location service
func NewLocationService(db *database.DB) *LocationService {
	return &LocationService{db: db}
}

// UpdateLocation updates a user's current location
func (s *LocationService) UpdateLocation(ctx context.Context, location *models.Location) error {
	location.ID = uuid.New().String()
	location.Timestamp = time.Now()
	location.CreatedAt = time.Now()
	location.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Create(location).Error
}

// GetCurrentLocation retrieves a user's current location
func (s *LocationService) GetCurrentLocation(ctx context.Context, userID, tenantID string) (*models.Location, error) {
	var location models.Location
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Order("timestamp DESC").
		First(&location).Error
	return &location, err
}

// GetLocationHistory retrieves a user's location history
func (s *LocationService) GetLocationHistory(ctx context.Context, userID, tenantID string, startTime, endTime time.Time) ([]models.Location, error) {
	var locations []models.Location
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?", 
			userID, tenantID, startTime, endTime).
		Order("timestamp ASC").
		Find(&locations).Error
	return locations, err
}

// SaveLocationHistory saves a batch of locations as history
func (s *LocationService) SaveLocationHistory(ctx context.Context, history *models.LocationHistory) error {
	history.ID = uuid.New().String()
	history.CreatedAt = time.Now()
	history.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Create(history).Error
}

// GetLocationStats retrieves location statistics for a user
func (s *LocationService) GetLocationStats(ctx context.Context, userID, tenantID string) (*models.LocationStats, error) {
	var stats models.LocationStats

	// Get total locations count
	if err := s.db.WithContext(ctx).Model(&models.Location{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Count(&stats.TotalLocations).Error; err != nil {
		return nil, err
	}

	// Get total history count
	if err := s.db.WithContext(ctx).Model(&models.LocationHistory{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Count(&stats.TotalHistory).Error; err != nil {
		return nil, err
	}

	// Get last location
	var lastLocation models.Location
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Order("timestamp DESC").
		First(&lastLocation).Error; err == nil {
		stats.LastLocation = lastLocation
		stats.LastUpdate = lastLocation.Timestamp
	}

	// Calculate average and max speed
	var speeds []float64
	if err := s.db.WithContext(ctx).Model(&models.Location{}).
		Where("user_id = ? AND tenant_id = ? AND speed > 0", userID, tenantID).
		Pluck("speed", &speeds).Error; err == nil {
		var totalSpeed float64
		for _, speed := range speeds {
			totalSpeed += speed
			if speed > stats.MaxSpeed {
				stats.MaxSpeed = speed
			}
		}
		if len(speeds) > 0 {
			stats.AverageSpeed = totalSpeed / float64(len(speeds))
		}
	}

	// Calculate total distance
	locations, err := s.GetLocationHistory(ctx, userID, tenantID, time.Now().Add(-24*time.Hour), time.Now())
	if err == nil {
		for i := 1; i < len(locations); i++ {
			stats.TotalDistance += calculateDistance(
				locations[i-1].Latitude, locations[i-1].Longitude,
				locations[i].Latitude, locations[i].Longitude,
			)
		}
	}

	return &stats, nil
}

// calculateDistance calculates the distance between two points using the Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth's radius in meters

	// Convert to radians
	lat1 = toRadians(lat1)
	lon1 = toRadians(lon1)
	lat2 = toRadians(lat2)
	lon2 = toRadians(lon2)

	// Haversine formula
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
} 