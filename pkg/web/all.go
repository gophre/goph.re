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

func All(c *gin.Context) {
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
		"articles": articles[:sizeNum],
		"all":      true,
	})
}
