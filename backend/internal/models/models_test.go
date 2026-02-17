package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBookFilter_SetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    BookFilter
		expected BookFilter
	}{
		{
			name:  "empty filter gets all defaults",
			input: BookFilter{},
			expected: BookFilter{
				Page:  1,
				Limit: 20,
				Sort:  "title",
				Order: "asc",
			},
		},
		{
			name:  "negative page corrected to 1",
			input: BookFilter{Page: -5, Limit: 50, Sort: "year", Order: "desc"},
			expected: BookFilter{
				Page:  1,
				Limit: 50,
				Sort:  "year",
				Order: "desc",
			},
		},
		{
			name:  "limit over 100 corrected to 20",
			input: BookFilter{Page: 3, Limit: 200, Sort: "title", Order: "asc"},
			expected: BookFilter{
				Page:  3,
				Limit: 20,
				Sort:  "title",
				Order: "asc",
			},
		},
		{
			name:  "valid filter unchanged",
			input: BookFilter{Page: 2, Limit: 50, Sort: "year", Order: "desc"},
			expected: BookFilter{
				Page:  2,
				Limit: 50,
				Sort:  "year",
				Order: "desc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.input
			f.SetDefaults()
			assert.Equal(t, tt.expected.Page, f.Page)
			assert.Equal(t, tt.expected.Limit, f.Limit)
			assert.Equal(t, tt.expected.Sort, f.Sort)
			assert.Equal(t, tt.expected.Order, f.Order)
		})
	}
}

func TestBookFilter_Offset(t *testing.T) {
	tests := []struct {
		page, limit, expected int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 50, 100},
		{1, 100, 0},
	}

	for _, tt := range tests {
		f := BookFilter{Page: tt.page, Limit: tt.limit}
		assert.Equal(t, tt.expected, f.Offset())
	}
}

func TestUser_ToUserInfo(t *testing.T) {
	now := time.Now()
	user := User{
		ID:           "uuid-123",
		Email:        "test@example.com",
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: "hashed_password",
		Role:         "admin",
		IsActive:     true,
		LastLoginAt:  &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	info := user.ToUserInfo()

	assert.Equal(t, "uuid-123", info.ID)
	assert.Equal(t, "test@example.com", info.Email)
	assert.Equal(t, "testuser", info.Username)
	assert.Equal(t, "Test User", info.DisplayName)
	assert.Equal(t, "admin", info.Role)
}
