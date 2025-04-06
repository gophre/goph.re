package rss

import (
	"encoding/json"
	"gophre/cmd/data"
	"gophre/env"
	"os"
	"strings"
)

func Search(query string, articles []data.Article, page, size int) []data.Article {
	var results []data.Article

	// Convert the query to lowercase for case-insensitive search
	query = strings.ToLower(query)

	// First collect all matching articles
	for _, article := range articles {
		if strings.Contains(strings.ToLower(article.Name), query) || strings.Contains(strings.ToLower(article.Resume), query) {
			results = append(results, article)
		}
	}

	// Then paginate the results
	startIdx := (page - 1) * size
	endIdx := startIdx + size

	// Check bounds
	if startIdx >= len(results) {
		return []data.Article{}
	}
	if endIdx > len(results) {
		endIdx = len(results)
	}

	return results[startIdx:endIdx]
}

// UpdateVoteByURL updates the vote status of an article identified by its URL
func UpdateVoteByURL(url string, vote string) bool {
	// Read the current articles from file
	file, err := os.ReadFile(env.POSTS)
	if err != nil {
		return false
	}

	var articles []data.Article
	err = json.Unmarshal(file, &articles)
	if err != nil {
		return false
	}

	// Update the vote for the article with matching URL
	articleFound := false
	for i := range articles {
		if articles[i].URL == url {
			articleFound = true
			if vote == "GOOD" {
				articles[i].Vote = 1
			} else if vote == "BAD" {
				articles[i].Vote = -1
			}
			break
		}
	}

	if !articleFound {
		return false
	}

	// Write the updated articles back to the file
	updatedFile, err := json.MarshalIndent(articles, "", "  ")
	if err != nil {
		return false
	}

	err = os.WriteFile(env.POSTS, updatedFile, 0644)
	return err == nil
}
