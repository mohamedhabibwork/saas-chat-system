package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"saas-chat-system/internal/models"
	"saas-chat-system/internal/services"
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

// @Summary      Create forum category
// @Description  Create a new forum category
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        category body models.ForumCategory true "Category details"
// @Success      201 {object} models.ForumCategory "Category created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/forum/categories [post]
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

// @Summary      Create forum topic
// @Description  Create a new forum topic
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        topic body models.ForumTopic true "Topic details"
// @Success      201 {object} models.ForumTopic "Topic created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/forum/topics [post]
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

// @Summary      Create forum post
// @Description  Create a new post in a forum topic
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        id path string true "Topic ID"
// @Param        post body models.ForumPost true "Post details"
// @Success      201 {object} models.ForumPost "Post created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Topic not found"
// @Router       /api/v1/forum/topics/{id}/posts [post]
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

// @Summary      Get forum categories
// @Description  Get all forum categories
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Success      200 {array} models.ForumCategory "List of categories"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/forum/categories [get]
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

// @Summary      Get topics in category
// @Description  Get all topics in a specific category
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        id path string true "Category ID"
// @Param        page query int false "Page number"
// @Param        limit query int false "Number of topics per page"
// @Success      200 {array} models.ForumTopic "List of topics"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Category not found"
// @Router       /api/v1/forum/categories/{id}/topics [get]
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

// @Summary      Get forum posts
// @Description  Get all posts in a forum topic
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        id path string true "Topic ID"
// @Param        page query int false "Page number"
// @Param        limit query int false "Number of posts per page"
// @Success      200 {array} models.ForumPost "List of posts"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Topic not found"
// @Router       /api/v1/forum/topics/{id}/posts [get]
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

// @Summary      Subscribe to topic
// @Description  Subscribe to notifications for a forum topic
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        id path string true "Topic ID"
// @Success      200 {object} map[string]interface{} "Subscribed successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Topic not found"
// @Router       /api/v1/forum/topics/{id}/subscribe [post]
func (h *ForumHandler) SubscribeToTopic(c *gin.Context) {
	topicID := c.Param("id")
	userID := c.MustGet("user_id").(string)

	if err := h.forumService.SubscribeToTopic(c.Request.Context(), topicID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscribed successfully"})
}

// @Summary      Unsubscribe from topic
// @Description  Unsubscribe from notifications for a forum topic
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        id path string true "Topic ID"
// @Success      200 {object} map[string]interface{} "Unsubscribed successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Topic not found"
// @Router       /api/v1/forum/topics/{id}/subscribe [delete]
func (h *ForumHandler) UnsubscribeFromTopic(c *gin.Context) {
	topicID := c.Param("id")
	userID := c.MustGet("user_id").(string)

	if err := h.forumService.UnsubscribeFromTopic(c.Request.Context(), topicID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed successfully"})
}

// @Summary      Update forum topic
// @Description  Update an existing forum topic
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        id path string true "Topic ID"
// @Param        topic body models.ForumTopic true "Updated topic details"
// @Success      200 {object} models.ForumTopic "Topic updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Topic not found"
// @Router       /api/v1/forum/topics/{id} [put]
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

// @Summary      Update forum post
// @Description  Update an existing forum post
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        id path string true "Post ID"
// @Param        post body models.ForumPost true "Updated post details"
// @Success      200 {object} models.ForumPost "Post updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Post not found"
// @Router       /api/v1/forum/posts/{id} [put]
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

// @Summary      Delete forum topic
// @Description  Delete an existing forum topic
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        id path string true "Topic ID"
// @Success      204 "Topic deleted successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Topic not found"
// @Router       /api/v1/forum/topics/{id} [delete]
func (h *ForumHandler) DeleteTopic(c *gin.Context) {
	topicID := c.Param("id")

	if err := h.forumService.DeleteTopic(c.Request.Context(), topicID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "topic deleted successfully"})
}

// @Summary      Delete forum post
// @Description  Delete an existing forum post
// @Tags         Forum
// @Accept       json
// @Produce      json
// @Param        id path string true "Post ID"
// @Success      204 "Post deleted successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Post not found"
// @Router       /api/v1/forum/posts/{id} [delete]
func (h *ForumHandler) DeletePost(c *gin.Context) {
	postID := c.Param("id")

	if err := h.forumService.DeletePost(c.Request.Context(), postID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}
