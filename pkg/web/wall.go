package web

import (
	"encoding/json"
	"fmt"
	"gophre/cmd/data"
	"gophre/env"
	"net/http"
	"os"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

func Wall(c *gin.Context) {
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

	// Create a new policy with bluemonday to sanitize the HTML
	p := bluemonday.UGCPolicy()

	// Show only the first 30 articles
	sizeNum := 30
	if sizeNum > len(articles) {
		sizeNum = len(articles)
	}

	for idx, article := range articles[:sizeNum] {
		// Sanitize the article content
		articles[idx].Resume = p.Sanitize(article.Resume)
	}

	c.HTML(http.StatusOK, "wall.html", gin.H{
		// "articles": articles[:sizeNum],
		"all": false,
	})
}

// TopArticles returns the most voted articles
func TopArticles(c *gin.Context) {
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

	// Filter articles with positive votes
	var goodArticles []data.Article
	for _, article := range articles {
		if article.Vote > 0 {
			goodArticles = append(goodArticles, article)
		}
	}

	// Sort by votes (highest first)
	sort.Slice(goodArticles, func(i, j int) bool {
		return goodArticles[i].Vote > goodArticles[j].Vote
	})

	// Limit to top 50
	maxArticles := 50
	if len(goodArticles) > maxArticles {
		goodArticles = goodArticles[:maxArticles]
	}

	// Sanitize content
	p := bluemonday.UGCPolicy()
	for i := range goodArticles {
		goodArticles[i].Resume = p.Sanitize(goodArticles[i].Resume)
	}

	c.HTML(http.StatusOK, "list.html", gin.H{
		"articles": goodArticles,
		"title":    "Top Voted Articles",
		"isEmpty":  len(goodArticles) == 0,
	})
}
