package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"saas-chat-system/internal/models"
)

// ForumService handles forum-related operations
type ForumService struct {
	db              *sql.DB
	notificationSvc *NotificationService
}

// NewForumService creates a new forum service
func NewForumService(db *sql.DB, notificationSvc *NotificationService) *ForumService {
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

	query := `
		INSERT INTO forum_categories (id, name, description, tenant_id, order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description,
		category.TenantID, category.Order, category.CreatedAt, category.UpdatedAt)
	if err != nil {
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

	query := `
		INSERT INTO forum_topics (
			id, title, content, category_id, tenant_id, user_id, 
			is_pinned, is_locked, views, created_at, updated_at, last_post_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := s.db.ExecContext(ctx, query,
		topic.ID, topic.Title, topic.Content, topic.CategoryID,
		topic.TenantID, topic.UserID, topic.IsPinned, topic.IsLocked,
		0, topic.CreatedAt, topic.UpdatedAt, topic.LastPostAt)
	if err != nil {
		return fmt.Errorf("error creating topic: %v", err)
	}

	return nil
}

// CreatePost creates a new post in a topic
func (s *ForumService) CreatePost(ctx context.Context, post *models.ForumPost) error {
	post.ID = uuid.New().String()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	// Insert the post
	query := `
		INSERT INTO forum_posts (
			id, topic_id, user_id, content, is_edited,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.ExecContext(ctx, query,
		post.ID, post.TopicID, post.UserID, post.Content,
		false, post.CreatedAt, post.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error creating post: %v", err)
	}

	// Update topic's last post time
	updateQuery := `
		UPDATE forum_topics
		SET last_post_at = $1
		WHERE id = $2
	`
	_, err = s.db.ExecContext(ctx, updateQuery, time.Now(), post.TopicID)
	if err != nil {
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
	subscribersQuery := `
		SELECT id, topic_id, user_id, created_at
		FROM forum_subscriptions
		WHERE topic_id = $1
	`
	rows, err := s.db.QueryContext(ctx, subscribersQuery, post.TopicID)
	if err != nil {
		return fmt.Errorf("error getting topic subscribers: %v", err)
	}
	defer rows.Close()

	// Process subscribers
	var subscribers []models.ForumSubscription
	for rows.Next() {
		var sub models.ForumSubscription
		err := rows.Scan(&sub.ID, &sub.TopicID, &sub.UserID, &sub.CreatedAt)
		if err != nil {
			return fmt.Errorf("error scanning subscription: %v", err)
		}
		subscribers = append(subscribers, sub)
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
	query := `
		SELECT id, name, description, tenant_id, order, created_at, updated_at
		FROM forum_categories
		WHERE tenant_id = $1
		ORDER BY "order" ASC
	`
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving categories: %v", err)
	}
	defer rows.Close()

	var categories []*models.ForumCategory
	for rows.Next() {
		var cat models.ForumCategory
		err := rows.Scan(
			&cat.ID, &cat.Name, &cat.Description,
			&cat.TenantID, &cat.Order, &cat.CreatedAt, &cat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning category: %v", err)
		}
		categories = append(categories, &cat)
	}

	return categories, nil
}

// GetTopics retrieves topics in a category
func (s *ForumService) GetTopics(ctx context.Context, categoryID string, page, limit int) ([]*models.ForumTopic, error) {
	offset := (page - 1) * limit
	query := `
		SELECT id, title, content, category_id, tenant_id, user_id, 
		       is_pinned, is_locked, views, created_at, updated_at, last_post_at
		FROM forum_topics
		WHERE category_id = $1
		ORDER BY is_pinned DESC, last_post_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.QueryContext(ctx, query, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving topics: %v", err)
	}
	defer rows.Close()

	var topics []*models.ForumTopic
	for rows.Next() {
		var topic models.ForumTopic
		err := rows.Scan(
			&topic.ID, &topic.Title, &topic.Content, &topic.CategoryID,
			&topic.TenantID, &topic.UserID, &topic.IsPinned, &topic.IsLocked,
			&topic.Views, &topic.CreatedAt, &topic.UpdatedAt, &topic.LastPostAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning topic: %v", err)
		}
		topics = append(topics, &topic)
	}

	return topics, nil
}

// GetPosts retrieves posts in a topic
func (s *ForumService) GetPosts(ctx context.Context, topicID string, page, limit int) ([]*models.ForumPost, error) {
	offset := (page - 1) * limit
	query := `
		SELECT id, topic_id, user_id, content, is_edited, created_at, updated_at
		FROM forum_posts
		WHERE topic_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.QueryContext(ctx, query, topicID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving posts: %v", err)
	}
	defer rows.Close()

	var posts []*models.ForumPost
	for rows.Next() {
		var post models.ForumPost
		err := rows.Scan(
			&post.ID, &post.TopicID, &post.UserID, &post.Content,
			&post.IsEdited, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning post: %v", err)
		}
		posts = append(posts, &post)
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

	query := `
		INSERT INTO forum_subscriptions (id, topic_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := s.db.ExecContext(ctx, query,
		subscription.ID, subscription.TopicID, subscription.UserID, subscription.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating subscription: %v", err)
	}

	return nil
}

// UnsubscribeFromTopic unsubscribes a user from a topic
func (s *ForumService) UnsubscribeFromTopic(ctx context.Context, topicID, userID string) error {
	query := `
		DELETE FROM forum_subscriptions
		WHERE topic_id = $1 AND user_id = $2
	`
	_, err := s.db.ExecContext(ctx, query, topicID, userID)
	if err != nil {
		return fmt.Errorf("error deleting subscription: %v", err)
	}

	return nil
}

// UpdateTopic updates a forum topic
func (s *ForumService) UpdateTopic(ctx context.Context, topic *models.ForumTopic) error {
	topic.UpdatedAt = time.Now()

	query := `
		UPDATE forum_topics
		SET title = $1, content = $2, is_pinned = $3, is_locked = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := s.db.ExecContext(ctx, query,
		topic.Title, topic.Content, topic.IsPinned, topic.IsLocked, topic.UpdatedAt, topic.ID)
	if err != nil {
		return fmt.Errorf("error updating topic: %v", err)
	}

	return nil
}

// UpdatePost updates a forum post
func (s *ForumService) UpdatePost(ctx context.Context, post *models.ForumPost) error {
	post.UpdatedAt = time.Now()
	post.IsEdited = true

	query := `
		UPDATE forum_posts
		SET content = $1, is_edited = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := s.db.ExecContext(ctx, query,
		post.Content, post.IsEdited, post.UpdatedAt, post.ID)
	if err != nil {
		return fmt.Errorf("error updating post: %v", err)
	}

	return nil
}

// DeleteTopic deletes a forum topic and all its posts
func (s *ForumService) DeleteTopic(ctx context.Context, topicID string) error {
	// Delete all posts first
	postsQuery := `DELETE FROM forum_posts WHERE topic_id = $1`
	_, err := s.db.ExecContext(ctx, postsQuery, topicID)
	if err != nil {
		return fmt.Errorf("error deleting posts: %v", err)
	}

	// Delete all subscriptions
	subscriptionsQuery := `DELETE FROM forum_subscriptions WHERE topic_id = $1`
	_, err = s.db.ExecContext(ctx, subscriptionsQuery, topicID)
	if err != nil {
		return fmt.Errorf("error deleting subscriptions: %v", err)
	}

	// Delete the topic
	topicQuery := `DELETE FROM forum_topics WHERE id = $1`
	_, err = s.db.ExecContext(ctx, topicQuery, topicID)
	if err != nil {
		return fmt.Errorf("error deleting topic: %v", err)
	}

	return nil
}

// DeletePost deletes a forum post
func (s *ForumService) DeletePost(ctx context.Context, postID string) error {
	query := `DELETE FROM forum_posts WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("error deleting post: %v", err)
	}

	return nil
}

// IncrementTopicViews increments the view count of a topic
func (s *ForumService) IncrementTopicViews(ctx context.Context, topicID string) error {
	query := `UPDATE forum_topics SET views = views + 1 WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, topicID)
	if err != nil {
		return fmt.Errorf("error incrementing views: %v", err)
	}

	return nil
}
