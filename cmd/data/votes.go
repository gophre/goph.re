package data

import "time"

type Vote struct {
	ArticleID string    `json:"article_id"`
	Value     int       `json:"value"`
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}