package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gophre/cmd/data"
	"gophre/env"
	"log"
	"os"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
)

func Update() {
	// Charger les sources RSS à partir d'un fichier JSON
	file, err := os.ReadFile(env.FEEDS)
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier :", err)
		return
	}

	var sources []data.Rss
	err = json.Unmarshal(file, &sources)
	if err != nil {
		fmt.Println("Erreur lors de la décodage du JSON :", err)
		return
	}

	// Récupérer les articles de chaque source RSS
	for _, src := range sources {
		fmt.Printf("%s:\n", src.Name)
		readSource(src)
		/*
			feed, err := fp.ParseURL(src.URL)
			if err != nil {
				fmt.Printf("Erreur lors de la récupération du flux %s: %v\n", src.Name, err)
				continue
			}

			for _, item := range feed.Items {
				timetest := *item.PublishedParsed
				articles = append(articles, data.Article{
					Source: src,
					Name:   item.Title,
					URL:    item.Link,
					Resume: item.Description,
					Date:   timetest,
					Vote:   "",
				})
			}
		*/
	}

	/*
		// Trier les articles par date
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Date.After(articles[j].Date)
		})

		// Sauvegarder les articles dans un fichier JSON
		jsonData, err := json.MarshalIndent(articles, "", "  ")
		if err != nil {
			fmt.Println("Erreur lors de l'encodage des articles en JSON :", err)
			return
		}

		err = os.WriteFile(env.POSTS, jsonData, 0644)
		if err != nil {
			fmt.Println("Erreur lors de l'écriture du fichier rss_posts.json :", err)
			return
		}
	*/
}

func parseDate(dateString string, layouts []string) (time.Time, error) {
	var parsedDate time.Time
	var err error

	for _, layout := range layouts {
		parsedDate, err = time.Parse(layout, dateString)
		if err == nil {
			return parsedDate, nil
		}
	}

	return parsedDate, err
}

func readSource(source data.Rss) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(source.URL)
	if err != nil {
		log.Printf("Error parsing URL %s: %v\n", source.URL, err)
		return
	}

	// Load existing articles from rss_posts.json
	file, err := os.ReadFile(env.POSTS)
	if err != nil {
		log.Printf("Error reading file: %v\n", err)
		return
	}

	var existingArticles []data.Article
	var newArticles []data.Article
	err = json.Unmarshal(file, &existingArticles)
	if err != nil {
		log.Printf("Error decoding JSON: %v\n", err)
		return
	}

	// Create a map to store URLs of existing articles
	existingURLs := make(map[string]bool)
	for _, article := range existingArticles {
		existingURLs[article.URL] = true
	}

	customRFC := "Mon, 2 Jan 2006 15:04:05"
	customRFC1123Z := "Mon, 2 Jan 2006 15:04:05 -0700"
	layouts := []string{
		customRFC,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		customRFC1123Z,
	}

	for _, item := range feed.Items {
		if !existingURLs[item.Link] { // Check if the article is new
			published, err := parseDate(item.Published, layouts)
			if err != nil {
				log.Printf("Error parsing published date: %v\n", err)
				continue
			}
			hash := md5.Sum([]byte(item.Title + "." + item.Link))
			hashInHex := hex.EncodeToString(hash[:])

			newArticle := data.Article{
				ID:     hashInHex,
				Source: source,
				Name:   item.Title,
				URL:    item.Link,
				Resume: item.Description,
				Date:   published,
			}
			newArticles = append(newArticles, newArticle)
		}
	}

	// If there are new articles, append them to existingArticles and update rss_posts.json
	if len(newArticles) > 0 {
		allArticles := append(existingArticles, newArticles...)
		sort.SliceStable(allArticles, func(i, j int) bool {
			return allArticles[i].Date.After(allArticles[j].Date)
		})

		updatedFile, err := json.MarshalIndent(allArticles, "", "  ")
		if err != nil {
			log.Printf("Error encoding JSON: %v\n", err)
			return
		}

		err = os.WriteFile(env.POSTS, updatedFile, 0644)
		if err != nil {
			log.Printf("Error writing file: %v\n", err)
			return
		}
	}
}
