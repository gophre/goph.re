package web

import (
	"encoding/json"
	"fmt"
	"gophre/cmd/data"
	"gophre/env"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

// GetUserBookmarks retrieves articles that a user has upvoted
func GetUserBookmarks(userID string) ([]data.Article, error) {
	// Read all votes
	var votes []data.Vote
	voteFile, err := os.ReadFile(env.VOTES)
	if err != nil {
		return nil, fmt.Errorf("error reading votes file: %v", err)
	}

	err = json.Unmarshal(voteFile, &votes)
	if err != nil {
		return nil, fmt.Errorf("error decoding votes: %v", err)
	}

	// Collect IDs of upvoted articles
	upvotedIDs := make(map[string]bool)
	for _, vote := range votes {
		if vote.UserID == userID && vote.Value > 0 {
			upvotedIDs[vote.ArticleID] = true
		}
	}

	// Read all articles
	articlesFile, err := os.ReadFile(env.POSTS)
	if err != nil {
		return nil, fmt.Errorf("error reading articles file: %v", err)
	}

	var allArticles []data.Article
	err = json.Unmarshal(articlesFile, &allArticles)
	if err != nil {
		return nil, fmt.Errorf("error decoding articles: %v", err)
	}

	// Filter to only include upvoted articles
	var bookmarkedArticles []data.Article
	for _, article := range allArticles {
		if upvotedIDs[article.ID] {
			bookmarkedArticles = append(bookmarkedArticles, article)
		}
	}

	return bookmarkedArticles, nil
}

// UserBookmarks handles the endpoint to show a user's bookmarked articles
func UserBookmarks(c *gin.Context) {
	user, isAuthenticated := GetCurrentUser(c)
	if !isAuthenticated {
		c.Redirect(http.StatusFound, "/auth/github")
		return
	}

	articles, err := GetUserBookmarks(user.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error retrieving bookmarks: %v", err))
		return
	}

	// Sanitize content
	p := bluemonday.UGCPolicy()
	for i := range articles {
		articles[i].Resume = p.Sanitize(articles[i].Resume)
	}

	c.HTML(http.StatusOK, "list.html", gin.H{
		"articles":  articles,
		"bookmarks": true,
		"user":      user,
	})
}

// LegacyBookmarks is kept for backward compatibility
func LegacyBookmarks(c *gin.Context) {
	file, err := os.ReadFile(env.POSTS)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading file: %v", err))
		return
	}

	var articles []data.Article
	err = json.Unmarshal(file, &articles)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error decoding JSON: %v", err))
		return
	}

	p := bluemonday.UGCPolicy()
	// Load all articles for JS to filter by localStorage
	for i := range articles {
		articles[i].Resume = p.Sanitize(articles[i].Resume)
	}

	c.HTML(http.StatusOK, "list.html", gin.H{
		"articles":         articles,
		"isEmpty":          len(articles) == 0,
		"useLocalStorage":  true,
		"localStorageOnly": true,
	})
}
