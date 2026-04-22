package domain

import "time"

type NewsStatus string

const (
	NewsDraft     NewsStatus = "draft"
	NewsPublished NewsStatus = "published"
	NewsArchived  NewsStatus = "archived"
)

type NewsArticle struct {
	ID        UUID       `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	Content   string     `json:"content" db:"content"`
	ImageURL  string     `json:"imageUrl" db:"image_url"`
	Status    NewsStatus `json:"status" db:"status"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}
