package rss

import (
	"encoding/json"
	"gophre/cmd/data"
	"gophre/env"
	"os"
	"strconv"

	"github.com/microcosm-cc/bluemonday"
)

func Feed(path, page, size string) (goodArticles []data.Article) {

	var articles []data.Article

	pageNum, err := strconv.Atoi(page)
	if err == nil {

		sizeNum, err := strconv.Atoi(size)
		if err == nil {
			file, err := os.ReadFile(env.POSTS)
			if err == nil {
				err = json.Unmarshal(file, &articles)
				if err == nil {

					for _, article := range articles {
						if path == "" || article.Source.Path == path {
							goodArticles = append(goodArticles, article)
						}
					}

					p := bluemonday.UGCPolicy()
					for idx, article := range goodArticles {
						goodArticles[idx].Resume = p.Sanitize(article.Resume)
					}

					startIndex := (pageNum - 1) * sizeNum
					endIndex := pageNum * sizeNum
					if startIndex >= len(goodArticles) {
						goodArticles = []data.Article{}
					}
					if endIndex > len(goodArticles) {
						endIndex = len(goodArticles)
					}
					return goodArticles[startIndex:endIndex]
				}
			}
		}
	}
	return []data.Article{}

}
