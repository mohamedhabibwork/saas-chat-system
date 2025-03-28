package handlers

import (
	"context"
	"fmt"
)

// getUserIDFromContext retrieves the user ID from the request context
func getUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value("user_id").(int)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
} 