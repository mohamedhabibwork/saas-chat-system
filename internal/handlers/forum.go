package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
	"github.com/mohamedhabibwork/saas-chat-system/internal/services"
)

// ForumHandler handles forum-related HTTP requests
type ForumHandler struct {
	forumService *services.ForumService
}

// NewForumHandler creates a new forum handler
func NewForumHandler(forumService *services.ForumService) *ForumHandler {
	return &ForumHandler{
		forumService: forumService,
	}
}

// CreateCategory handles requests to create a new forum category
func (h *ForumHandler) CreateCategory(c *gin.Context) {
	var category models.ForumCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}
	category.TenantID = tenantID.(string)

	if err := h.forumService.CreateCategory(c.Request.Context(), &category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// CreateTopic handles requests to create a new forum topic
func (h *ForumHandler) CreateTopic(c *gin.Context) {
	var topic models.ForumTopic
	if err := c.ShouldBindJSON(&topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant and user IDs from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}
	topic.TenantID = tenantID.(string)

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	topic.UserID = userID.(string)

	if err := h.forumService.CreateTopic(c.Request.Context(), &topic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, topic)
}

// CreatePost handles requests to create a new post in a topic
func (h *ForumHandler) CreatePost(c *gin.Context) {
	topicID := c.Param("id")
	var post models.ForumPost
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	post.UserID = userID.(string)
	post.TopicID = topicID

	if err := h.forumService.CreatePost(c.Request.Context(), &post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// GetCategories handles requests to retrieve forum categories
func (h *ForumHandler) GetCategories(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}

	categories, err := h.forumService.GetCategories(c.Request.Context(), tenantID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetTopics handles requests to retrieve topics in a category
func (h *ForumHandler) GetTopics(c *gin.Context) {
	categoryID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	topics, err := h.forumService.GetTopics(c.Request.Context(), categoryID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, topics)
}

// GetPosts handles requests to retrieve posts in a topic
func (h *ForumHandler) GetPosts(c *gin.Context) {
	topicID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	posts, err := h.forumService.GetPosts(c.Request.Context(), topicID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// SubscribeToTopic handles requests to subscribe to a topic
func (h *ForumHandler) SubscribeToTopic(c *gin.Context) {
	topicID := c.Param("id")
	userID := c.MustGet("user_id").(string)

	if err := h.forumService.SubscribeToTopic(c.Request.Context(), topicID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscribed successfully"})
}

// UnsubscribeFromTopic handles requests to unsubscribe from a topic
func (h *ForumHandler) UnsubscribeFromTopic(c *gin.Context) {
	topicID := c.Param("id")
	userID := c.MustGet("user_id").(string)

	if err := h.forumService.UnsubscribeFromTopic(c.Request.Context(), topicID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed successfully"})
}

// UpdateTopic handles requests to update a forum topic
func (h *ForumHandler) UpdateTopic(c *gin.Context) {
	topicID := c.Param("id")
	var topic models.ForumTopic
	if err := c.ShouldBindJSON(&topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic.ID = topicID
	if err := h.forumService.UpdateTopic(c.Request.Context(), &topic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, topic)
}

// UpdatePost handles requests to update a forum post
func (h *ForumHandler) UpdatePost(c *gin.Context) {
	postID := c.Param("id")
	var post models.ForumPost
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.ID = postID
	if err := h.forumService.UpdatePost(c.Request.Context(), &post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// DeleteTopic handles requests to delete a forum topic
func (h *ForumHandler) DeleteTopic(c *gin.Context) {
	topicID := c.Param("id")

	if err := h.forumService.DeleteTopic(c.Request.Context(), topicID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "topic deleted successfully"})
}

// DeletePost handles requests to delete a forum post
func (h *ForumHandler) DeletePost(c *gin.Context) {
	postID := c.Param("id")

	if err := h.forumService.DeletePost(c.Request.Context(), postID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
} 