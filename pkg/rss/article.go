package rss

import (
	"encoding/json"
	"gophre/cmd/data"
	"gophre/env"
	"os"

	"github.com/microcosm-cc/bluemonday"
)

func Article(uid string) (post data.Article) {
	file, err := os.ReadFile(env.POSTS)
	if err == nil {
		var articles []data.Article
		err = json.Unmarshal(file, &articles)
		if err == nil {
			p := bluemonday.UGCPolicy()
			for _, article := range articles {
				if article.ID == uid {
					post = article
					post.Resume = p.Sanitize(article.Resume)
					break
				}
			}
		}
	}

	return
}
