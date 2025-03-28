package services

import (
	"context"
	"database/sql"
	"errors"
)

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (email, username, password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	
	err := s.db.QueryRowContext(ctx, query,
		user.Email,
		user.Username,
		user.Password,
		user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return err
	}
	
	return nil
}

func (s *UserService) Get(ctx context.Context, id string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, username, role, created_at, updated_at
		FROM users
		WHERE id = $1`
	
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (s *UserService) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET email = $1, username = $2, role = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at`
	
	err := s.db.QueryRowContext(ctx, query,
		user.Email,
		user.Username,
		user.Role,
		user.ID,
	).Scan(&user.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}
	if err != nil {
		return err
	}
	
	return nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	
	return nil
}

func (s *UserService) List(ctx context.Context, page, limit int) ([]*User, error) {
	query := `
		SELECT id, email, username, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	
	offset := (page - 1) * limit
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return users, nil
} 