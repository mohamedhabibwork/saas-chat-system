package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedhabibwork/saas-chat-system/internal/database"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
)

// ForumService handles forum-related operations
type ForumService struct {
	db              *database.DB
	notificationSvc *NotificationService
}

// NewForumService creates a new forum service
func NewForumService(db *database.DB, notificationSvc *NotificationService) *ForumService {
	return &ForumService{
		db:              db,
		notificationSvc: notificationSvc,
	}
}

// CreateCategory creates a new forum category
func (s *ForumService) CreateCategory(ctx context.Context, category *models.ForumCategory) error {
	category.ID = uuid.New().String()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Create(category).Error; err != nil {
		return fmt.Errorf("error creating category: %v", err)
	}

	return nil
}

// CreateTopic creates a new forum topic
func (s *ForumService) CreateTopic(ctx context.Context, topic *models.ForumTopic) error {
	topic.ID = uuid.New().String()
	topic.CreatedAt = time.Now()
	topic.UpdatedAt = time.Now()
	topic.LastPostAt = time.Now()

	if err := s.db.WithContext(ctx).Create(topic).Error; err != nil {
		return fmt.Errorf("error creating topic: %v", err)
	}

	return nil
}

// CreatePost creates a new post in a topic
func (s *ForumService) CreatePost(ctx context.Context, post *models.ForumPost) error {
	post.ID = uuid.New().String()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Create(post).Error; err != nil {
		return fmt.Errorf("error creating post: %v", err)
	}

	// Update topic's last post time
	if err := s.db.WithContext(ctx).Model(&models.ForumTopic{}).Where("id = ?", post.TopicID).
		Update("last_post_at", time.Now()).Error; err != nil {
		return fmt.Errorf("error updating topic last post time: %v", err)
	}

	// Create notification for topic subscribers
	notification := &models.ForumNotification{
		ID:        uuid.New().String(),
		TopicID:   post.TopicID,
		PostID:    post.ID,
		Type:      "new_post",
		CreatedAt: time.Now(),
	}

	// Get topic subscribers
	var subscribers []models.ForumSubscription
	if err := s.db.WithContext(ctx).Where("topic_id = ?", post.TopicID).Find(&subscribers).Error; err != nil {
		return fmt.Errorf("error getting topic subscribers: %v", err)
	}

	// Send notifications to subscribers
	for _, subscriber := range subscribers {
		if subscriber.UserID != post.UserID { // Don't notify the post author
			notification.UserID = subscriber.UserID
			if err := s.notificationSvc.SendForumNotification(ctx, notification); err != nil {
				// Log error but continue with other notifications
				fmt.Printf("Error sending notification to user %s: %v\n", subscriber.UserID, err)
			}
		}
	}

	return nil
}

// GetCategories retrieves all forum categories
func (s *ForumService) GetCategories(ctx context.Context, tenantID string) ([]*models.ForumCategory, error) {
	var categories []*models.ForumCategory
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("order ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("error retrieving categories: %v", err)
	}
	return categories, nil
}

// GetTopics retrieves topics in a category
func (s *ForumService) GetTopics(ctx context.Context, categoryID string, page, limit int) ([]*models.ForumTopic, error) {
	var topics []*models.ForumTopic
	offset := (page - 1) * limit

	if err := s.db.WithContext(ctx).Where("category_id = ?", categoryID).
		Order("is_pinned DESC, last_post_at DESC").
		Offset(offset).Limit(limit).
		Find(&topics).Error; err != nil {
		return nil, fmt.Errorf("error retrieving topics: %v", err)
	}
	return topics, nil
}

// GetPosts retrieves posts in a topic
func (s *ForumService) GetPosts(ctx context.Context, topicID string, page, limit int) ([]*models.ForumPost, error) {
	var posts []*models.ForumPost
	offset := (page - 1) * limit

	if err := s.db.WithContext(ctx).Where("topic_id = ?", topicID).
		Order("created_at ASC").
		Offset(offset).Limit(limit).
		Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("error retrieving posts: %v", err)
	}
	return posts, nil
}

// SubscribeToTopic subscribes a user to a topic
func (s *ForumService) SubscribeToTopic(ctx context.Context, topicID, userID string) error {
	subscription := &models.ForumSubscription{
		ID:        uuid.New().String(),
		TopicID:   topicID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(subscription).Error; err != nil {
		return fmt.Errorf("error creating subscription: %v", err)
	}

	return nil
}

// UnsubscribeFromTopic unsubscribes a user from a topic
func (s *ForumService) UnsubscribeFromTopic(ctx context.Context, topicID, userID string) error {
	if err := s.db.WithContext(ctx).Where("topic_id = ? AND user_id = ?", topicID, userID).
		Delete(&models.ForumSubscription{}).Error; err != nil {
		return fmt.Errorf("error deleting subscription: %v", err)
	}

	return nil
}

// UpdateTopic updates a forum topic
func (s *ForumService) UpdateTopic(ctx context.Context, topic *models.ForumTopic) error {
	topic.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(topic).Error; err != nil {
		return fmt.Errorf("error updating topic: %v", err)
	}

	return nil
}

// UpdatePost updates a forum post
func (s *ForumService) UpdatePost(ctx context.Context, post *models.ForumPost) error {
	post.UpdatedAt = time.Now()
	post.IsEdited = true

	if err := s.db.WithContext(ctx).Save(post).Error; err != nil {
		return fmt.Errorf("error updating post: %v", err)
	}

	return nil
}

// DeleteTopic deletes a forum topic and all its posts
func (s *ForumService) DeleteTopic(ctx context.Context, topicID string) error {
	// Delete all posts first
	if err := s.db.WithContext(ctx).Where("topic_id = ?", topicID).
		Delete(&models.ForumPost{}).Error; err != nil {
		return fmt.Errorf("error deleting posts: %v", err)
	}

	// Delete all subscriptions
	if err := s.db.WithContext(ctx).Where("topic_id = ?", topicID).
		Delete(&models.ForumSubscription{}).Error; err != nil {
		return fmt.Errorf("error deleting subscriptions: %v", err)
	}

	// Delete the topic
	if err := s.db.WithContext(ctx).Delete(&models.ForumTopic{}, "id = ?", topicID).Error; err != nil {
		return fmt.Errorf("error deleting topic: %v", err)
	}

	return nil
}

// DeletePost deletes a forum post
func (s *ForumService) DeletePost(ctx context.Context, postID string) error {
	if err := s.db.WithContext(ctx).Delete(&models.ForumPost{}, "id = ?", postID).Error; err != nil {
		return fmt.Errorf("error deleting post: %v", err)
	}

	return nil
}

// IncrementTopicViews increments the view count of a topic
func (s *ForumService) IncrementTopicViews(ctx context.Context, topicID string) error {
	if err := s.db.WithContext(ctx).Model(&models.ForumTopic{}).
		Where("id = ?", topicID).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error; err != nil {
		return fmt.Errorf("error incrementing views: %v", err)
	}

	return nil
} 