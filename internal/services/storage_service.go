package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"awesomeProject/internal/config"
	"awesomeProject/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// StorageService handles file storage operations
type StorageService struct {
	db Database
	s3Client *s3.Client
	subscriptionService *SubscriptionService
}

// NewStorageService creates a new storage service
func NewStorageService(db Database, subscriptionService *SubscriptionService) (*StorageService, error) {
	cfg := config.GetConfig()
	var s3Client *s3.Client

	if cfg.Storage.Type == "s3" {
		// Configure AWS SDK
		awsCfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Storage.S3.Region),
			config.WithCredentials(credentials.NewStaticCredentialsProvider(
				cfg.Storage.S3.AccessKeyID,
				cfg.Storage.S3.SecretAccessKey,
				"",
			)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %v", err)
		}

		s3Client = s3.NewFromConfig(awsCfg)
	}

	return &StorageService{
		db: db,
		s3Client: s3Client,
		subscriptionService: subscriptionService,
	}, nil
}

// UploadFile uploads a file and returns its metadata
func (s *StorageService) UploadFile(userID int, file *multipart.FileHeader) (*models.File, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Get user's active subscription
	subscription, err := s.subscriptionService.GetActiveSubscription(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %v", err)
	}
	if subscription == nil {
		return nil, fmt.Errorf("no active subscription")
	}

	// Check subscription limits
	if err := s.checkStorageLimit(userID); err != nil {
		return nil, err
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// Generate unique filename
	filename := generateUniqueFilename(file.Filename)
	filepath := filepath.Join(config.GetUploadPath(userID), filename)

	// Upload based on storage type
	var fileURL string
	cfg := config.GetConfig()

	if cfg.Storage.Type == "local" {
		// Ensure directory exists
		if err := os.MkdirAll(filepath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %v", err)
		}

		// Create destination file
		dst, err := os.Create(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %v", err)
		}
		defer dst.Close()

		// Copy file content
		if _, err := io.Copy(dst, src); err != nil {
			return nil, fmt.Errorf("failed to copy file: %v", err)
		}

		fileURL = fmt.Sprintf("/files/%s", filepath)
	} else if cfg.Storage.Type == "s3" {
		// Upload to S3
		_, err := s.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(cfg.Storage.S3.Bucket),
			Key:    aws.String(filepath),
			Body:   src,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload to S3: %v", err)
		}

		fileURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
			cfg.Storage.S3.Bucket,
			cfg.Storage.S3.Region,
			filepath,
		)
	}

	// Create file record in database
	fileRecord := &models.File{
		UserID:      userID,
		Filename:    file.Filename,
		Filepath:    filepath,
		URL:         fileURL,
		Size:        file.Size,
		ContentType: file.Header.Get("Content-Type"),
	}

	if err := s.saveFileRecord(tx, fileRecord); err != nil {
		return nil, fmt.Errorf("failed to save file record: %v", err)
	}

	// Update subscription usage
	if err := s.subscriptionService.UpdateFileUsage(subscription.ID, file.Size); err != nil {
		return nil, fmt.Errorf("failed to update usage: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return fileRecord, nil
}

// DeleteFile deletes a file and its record
func (s *StorageService) DeleteFile(fileID int, userID int) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Get file record
	file, err := s.getFileRecord(fileID)
	if err != nil {
		return err
	}

	// Verify ownership
	if file.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Get user's active subscription
	subscription, err := s.subscriptionService.GetActiveSubscription(userID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %v", err)
	}
	if subscription == nil {
		return fmt.Errorf("no active subscription")
	}

	// Delete file based on storage type
	cfg := config.GetConfig()
	if cfg.Storage.Type == "local" {
		if err := os.Remove(file.Filepath); err != nil {
			return fmt.Errorf("failed to delete file: %v", err)
		}
	} else if cfg.Storage.Type == "s3" {
		_, err := s.s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
			Bucket: aws.String(cfg.Storage.S3.Bucket),
			Key:    aws.String(file.Filepath),
		})
		if err != nil {
			return fmt.Errorf("failed to delete from S3: %v", err)
		}
	}

	// Delete database record
	if err := s.deleteFileRecord(tx, fileID); err != nil {
		return fmt.Errorf("failed to delete file record: %v", err)
	}

	// Update subscription usage
	if err := s.subscriptionService.UpdateStorageUsage(subscription.ID, -file.Size); err != nil {
		return fmt.Errorf("failed to update usage: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
// GetFile retrieves a file
func (s *StorageService) GetFile(fileID int, userID int) (*models.File, io.ReadCloser, error) {
	// Get file record
	file, err := s.getFileRecord(fileID)
	if err != nil {
		return nil, nil, err
	}

	// Verify ownership
	if file.UserID != userID {
		return nil, nil, fmt.Errorf("unauthorized")
	}

	// Get file content based on storage type
	var reader io.ReadCloser
	cfg := config.GetConfig()

	if cfg.Storage.Type == "local" {
		reader, err = os.Open(file.Filepath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open file: %v", err)
		}
	} else if cfg.Storage.Type == "s3" {
		result, err := s.s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(cfg.Storage.S3.Bucket),
			Key:    aws.String(file.Filepath),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get from S3: %v", err)
		}
		reader = result.Body
	}

	return file, reader, nil
}

// ListFiles lists files for a user
func (s *StorageService) ListFiles(userID int) ([]models.File, error) {
	query := `
		SELECT id, user_id, filename, filepath, url, size, content_type, created_at
		FROM files
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		err := rows.Scan(
			&file.ID, &file.UserID, &file.Filename,
			&file.Filepath, &file.URL, &file.Size,
			&file.ContentType, &file.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// Helper functions

func (s *StorageService) checkStorageLimit(userID int) error {
	// Get user's subscription
	subscription, err := s.getUserSubscription(userID)
	if err != nil {
		return err
	}

	// Get current storage usage
	usage, err := s.getStorageUsage(userID)
	if err != nil {
		return err
	}

	// Check if usage exceeds limit
	if usage >= subscription.Plan.Limits.MaxStorageGB*1024*1024*1024 {
		return fmt.Errorf("storage limit exceeded")
	}

	return nil
}

func (s *StorageService) getUserSubscription(userID int) (*models.Subscription, error) {
	query := `
		SELECT s.id, s.user_id, s.plan_id, s.status, s.start_date, s.end_date,
			   s.auto_renew, s.payment_method, s.created_at, s.updated_at,
			   p.id, p.name, p.description, p.price, p.interval, p.features,
			   p.limits, p.created_at, p.updated_at
		FROM subscriptions s
		JOIN plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC
		LIMIT 1
	`
	var subscription models.Subscription
	err := s.db.QueryRow(query, userID).Scan(
		&subscription.ID, &subscription.UserID, &subscription.PlanID,
		&subscription.Status, &subscription.StartDate, &subscription.EndDate,
		&subscription.AutoRenew, &subscription.PaymentMethod,
		&subscription.CreatedAt, &subscription.UpdatedAt,
		&subscription.Plan.ID, &subscription.Plan.Name,
		&subscription.Plan.Description, &subscription.Plan.Price,
		&subscription.Plan.Interval, &subscription.Plan.Features,
		&subscription.Plan.Limits, &subscription.Plan.CreatedAt,
		&subscription.Plan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (s *StorageService) getStorageUsage(userID int) (int64, error) {
	query := `
		SELECT COALESCE(SUM(size), 0)
		FROM files
		WHERE user_id = $1
	`
	var usage int64
	err := s.db.QueryRow(query, userID).Scan(&usage)
	if err != nil {
		return 0, err
	}
	return usage, nil
}

func (s *StorageService) saveFileRecord(file *models.File) error {
	query := `
		INSERT INTO files (user_id, filename, filepath, url, size, content_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id
	`
	return s.db.QueryRow(
		query,
		file.UserID, file.Filename, file.Filepath,
		file.URL, file.Size, file.ContentType,
	).Scan(&file.ID)
}

func (s *StorageService) getFileRecord(fileID int) (*models.File, error) {
	query := `
		SELECT id, user_id, filename, filepath, url, size, content_type, created_at
		FROM files
		WHERE id = $1
	`
	var file models.File
	err := s.db.QueryRow(query, fileID).Scan(
		&file.ID, &file.UserID, &file.Filename,
		&file.Filepath, &file.URL, &file.Size,
		&file.ContentType, &file.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *StorageService) deleteFileRecord(fileID int) error {
	query := "DELETE FROM files WHERE id = $1"
	_, err := s.db.Exec(query, fileID)
	return err
}

func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := filepath.Base(originalFilename[:len(originalFilename)-len(ext)])
	return fmt.Sprintf("%s_%d%s", name, time.Now().UnixNano(), ext)
} 