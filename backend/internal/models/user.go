package models

import "time"

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	DisplayName  string     `json:"display_name"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func (u *User) ToUserInfo() UserInfo {
	return UserInfo{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Role:        u.Role,
	}
}

type CreateUserInput struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=200"`
	Password    string `json:"password" binding:"required,min=8,max=72"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ImportStats struct {
	BooksAdded   int   `json:"books_added"`
	BooksUpdated int   `json:"books_updated"`
	BooksDeleted int   `json:"books_deleted"`
	AuthorsAdded int   `json:"authors_added"`
	GenresAdded  int   `json:"genres_added"`
	SeriesAdded  int   `json:"series_added"`
	Errors       int   `json:"errors"`
	DurationMs   int64 `json:"duration_ms"`
}

type ImportStatus struct {
	Status         string       `json:"status"` // idle, running, completed, failed
	StartedAt      *time.Time   `json:"started_at,omitempty"`
	FinishedAt     *time.Time   `json:"finished_at,omitempty"`
	Stats          *ImportStats `json:"stats,omitempty"`
	Error          *string      `json:"error,omitempty"`
	TotalRecords   int          `json:"total_records,omitempty"`
	ProcessedBatch int          `json:"processed_batch,omitempty"`
	TotalBatches   int          `json:"total_batches,omitempty"`
}
