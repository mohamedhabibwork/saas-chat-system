package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"saas-chat-system/internal/models"
)

// LocationService handles location-related operations
type LocationService struct {
	db *sql.DB
}

// NewLocationService creates a new location service
func NewLocationService(db *sql.DB) *LocationService {
	return &LocationService{db: db}
}

// UpdateLocation updates a user's current location
func (s *LocationService) UpdateLocation(ctx context.Context, location *models.Location) error {
	location.ID = uuid.New().String()
	location.Timestamp = time.Now()
	location.CreatedAt = time.Now()
	location.UpdatedAt = time.Now()

	query := `
		INSERT INTO locations (
			id, user_id, tenant_id, latitude, longitude, 
			altitude, accuracy, speed, timestamp, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := s.db.ExecContext(ctx, query,
		location.ID, location.UserID, location.TenantID,
		location.Latitude, location.Longitude, location.Altitude,
		location.Accuracy, location.Speed, location.Timestamp,
		location.CreatedAt, location.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error updating location: %v", err)
	}

	return nil
}

// GetCurrentLocation retrieves a user's current location
func (s *LocationService) GetCurrentLocation(ctx context.Context, userID, tenantID string) (*models.Location, error) {
	query := `
		SELECT id, user_id, tenant_id, latitude, longitude, 
			   altitude, accuracy, speed, timestamp, created_at, updated_at
		FROM locations
		WHERE user_id = $1 AND tenant_id = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var location models.Location
	err := s.db.QueryRowContext(ctx, query, userID, tenantID).Scan(
		&location.ID, &location.UserID, &location.TenantID,
		&location.Latitude, &location.Longitude, &location.Altitude,
		&location.Accuracy, &location.Speed, &location.Timestamp,
		&location.CreatedAt, &location.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting current location: %v", err)
	}
	
	return &location, nil
}

// GetLocationHistory retrieves a user's location history
func (s *LocationService) GetLocationHistory(ctx context.Context, userID, tenantID string, startTime, endTime time.Time) ([]models.Location, error) {
	query := `
		SELECT id, user_id, tenant_id, latitude, longitude, 
			   altitude, accuracy, speed, timestamp, created_at, updated_at
		FROM locations
		WHERE user_id = $1 AND tenant_id = $2 AND timestamp BETWEEN $3 AND $4
		ORDER BY timestamp ASC
	`
	rows, err := s.db.QueryContext(ctx, query, userID, tenantID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("error getting location history: %v", err)
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		err := rows.Scan(
			&loc.ID, &loc.UserID, &loc.TenantID,
			&loc.Latitude, &loc.Longitude, &loc.Altitude,
			&loc.Accuracy, &loc.Speed, &loc.Timestamp,
			&loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning location: %v", err)
		}
		locations = append(locations, loc)
	}
	
	return locations, nil
}

// SaveLocationHistory saves a batch of locations as history
func (s *LocationService) SaveLocationHistory(ctx context.Context, history *models.LocationHistory) error {
	history.ID = uuid.New().String()
	history.CreatedAt = time.Now()
	history.UpdatedAt = time.Now()
	
	// Marshal locations to JSON for storage
	locationsJSON, err := json.Marshal(history.Locations)
	if err != nil {
		return fmt.Errorf("error marshaling location data: %v", err)
	}

	query := `
		INSERT INTO location_history (
			id, user_id, tenant_id, data, start_time, 
			end_time, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = s.db.ExecContext(ctx, query,
		history.ID, history.UserID, history.TenantID,
		locationsJSON, history.StartTime, history.EndTime,
		history.CreatedAt, history.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error saving location history: %v", err)
	}

	return nil
}

// GetLocationStats retrieves location statistics for a user
func (s *LocationService) GetLocationStats(ctx context.Context, userID, tenantID string) (*models.LocationStats, error) {
	var stats models.LocationStats

	// Get total locations count
	countQuery := `SELECT COUNT(*) FROM locations WHERE user_id = $1 AND tenant_id = $2`
	err := s.db.QueryRowContext(ctx, countQuery, userID, tenantID).Scan(&stats.TotalLocations)
	if err != nil {
		return nil, fmt.Errorf("error counting locations: %v", err)
	}

	// Get total history count
	historyCountQuery := `SELECT COUNT(*) FROM location_history WHERE user_id = $1 AND tenant_id = $2`
	err = s.db.QueryRowContext(ctx, historyCountQuery, userID, tenantID).Scan(&stats.TotalHistory)
	if err != nil {
		return nil, fmt.Errorf("error counting history: %v", err)
	}

	// Get last location
	lastLocationQuery := `
		SELECT id, user_id, tenant_id, latitude, longitude, 
			   altitude, accuracy, speed, timestamp, created_at, updated_at
		FROM locations
		WHERE user_id = $1 AND tenant_id = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var lastLocation models.Location
	err = s.db.QueryRowContext(ctx, lastLocationQuery, userID, tenantID).Scan(
		&lastLocation.ID, &lastLocation.UserID, &lastLocation.TenantID,
		&lastLocation.Latitude, &lastLocation.Longitude, &lastLocation.Altitude,
		&lastLocation.Accuracy, &lastLocation.Speed, &lastLocation.Timestamp,
		&lastLocation.CreatedAt, &lastLocation.UpdatedAt,
	)
	if err == nil {
		stats.LastLocation = lastLocation
		stats.LastUpdate = lastLocation.Timestamp
	}

	// Calculate average and max speed
	speedQuery := `
		SELECT speed FROM locations
		WHERE user_id = $1 AND tenant_id = $2 AND speed > 0
	`
	rows, err := s.db.QueryContext(ctx, speedQuery, userID, tenantID)
	if err == nil {
		defer rows.Close()
		
		var speeds []float64
		for rows.Next() {
			var speed float64
			if err := rows.Scan(&speed); err == nil {
				speeds = append(speeds, speed)
				if speed > stats.MaxSpeed {
					stats.MaxSpeed = speed
				}
			}
		}
		
		if len(speeds) > 0 {
			var totalSpeed float64
			for _, speed := range speeds {
				totalSpeed += speed
			}
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
