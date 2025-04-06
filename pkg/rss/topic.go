package rss

import (
	"encoding/json"
	"gophre/cmd/data"
	"gophre/env"
	"os"

	"github.com/microcosm-cc/bluemonday"
)

func Topic(path string) []data.Article {
	file, err := os.ReadFile(env.POSTS)
	if err == nil {

		var articles []data.Article
		var goodArticles []data.Article
		err = json.Unmarshal(file, &articles)
		if err == nil {

			for _, article := range articles {
				if article.Source.Path == path {
					goodArticles = append(goodArticles, article)
				}
			}

			// Create a new policy with bluemonday to sanitize the HTML
			p := bluemonday.UGCPolicy()

			// Show only the first 30 articles
			sizeNum := 30
			if sizeNum > len(goodArticles) {
				sizeNum = len(goodArticles)
			}

			for idx, article := range goodArticles[:sizeNum] {
				// Sanitize the article content
				articles[idx].Resume = p.Sanitize(article.Resume)
			}

			return goodArticles[:sizeNum]
		}
	}

	return []data.Article{}
}
