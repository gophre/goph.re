package data

import "time"

type Article struct {
	ID     string    `json:"id"`
	Source Rss       `json:"source"`
	Name   string    `json:"name"`
	URL    string    `json:"url"`
	Resume string    `json:"resume"`
	Date   time.Time `json:"date"`
	Vote   int       `json:"vote"`
}
