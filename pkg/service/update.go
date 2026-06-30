package service

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gophre/cmd/data"
	"gophre/env"
	"html"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
)

const (
	// discordWebhookURL is the Discord webhook called for each new article.
	discordWebhookURL = "https://discord.com/api/webhooks/1521619888190263367/S4Sadt3tEEIz0BnFuKmQOsWBDWY1-miRXa9oHMRlNWT9A20Gh33fZAIG0GUmvuAcE-zz"

	// discordContentLimit is Discord's hard cap on a message's content field.
	discordContentLimit = 2000

	// discordMinInterval throttles webhook calls to stay under Discord's rate limit.
	discordMinInterval = 1500 * time.Millisecond

	// notifyMaxAge limits Discord notifications to recently published articles so a
	// backfill of old feed items does not flood the channel.
	notifyMaxAge = 7 * 24 * time.Hour
)

// discordStripHTML removes all markup from feed descriptions before posting.
var discordStripHTML = bluemonday.StrictPolicy()

// lastDiscordCall tracks the last webhook call for simple client-side throttling.
var lastDiscordCall time.Time

// truncate shortens s to at most n runes, appending an ellipsis when cut.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

// plainText strips all HTML markup and entities from s, returning collapsed
// raw text suitable for a Discord message.
func plainText(s string) string {
	// Strip tags, then unescape entities, then strip again in case an entity
	// such as "&lt;b&gt;" decoded into a real tag.
	s = discordStripHTML.Sanitize(s)
	s = html.UnescapeString(s)
	s = discordStripHTML.Sanitize(s)
	s = html.UnescapeString(s)
	return strings.Join(strings.Fields(s), " ")
}

// notifyDiscord posts the given article to the Discord webhook, sanitizing and
// truncating the content and respecting Discord's rate limits.
func notifyDiscord(article data.Article) {
	title := plainText(article.Name)
	resume := plainText(article.Resume)

	var b strings.Builder
	if title != "" {
		b.WriteString("**" + title + "**\n")
	}
	if resume != "" {
		b.WriteString(resume + "\n")
	}
	b.WriteString(article.URL)

	content := truncate(strings.TrimSpace(b.String()), discordContentLimit)
	if content == "" {
		return // Nothing meaningful to post; avoids a guaranteed 400.
	}

	payload, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		log.Printf("Error encoding Discord payload: %v\n", err)
		return
	}

	// Up to 5 attempts so a 429 can be retried after the advised delay.
	for attempt := 0; attempt < 5; attempt++ {
		if wait := discordMinInterval - time.Since(lastDiscordCall); wait > 0 {
			time.Sleep(wait)
		}

		resp, err := http.Post(discordWebhookURL, "application/json", bytes.NewReader(payload))
		lastDiscordCall = time.Now()
		if err != nil {
			log.Printf("Error calling Discord webhook: %v\n", err)
			return
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			wait := discordRetryAfter(resp)
			resp.Body.Close()
			log.Printf("Discord rate limited, retrying in %s\n", wait)
			time.Sleep(wait)
			continue
		}

		if resp.StatusCode >= 300 {
			log.Printf("Discord webhook returned status: %s\n", resp.Status)
		}
		resp.Body.Close()
		return
	}

	log.Printf("Discord webhook gave up after repeated rate limiting for: %s\n", article.URL)
}

// discordRetryAfter reads the retry delay advised by a 429 response, falling
// back to the client throttle interval when no hint is provided.
func discordRetryAfter(resp *http.Response) time.Duration {
	if v := resp.Header.Get("Retry-After"); v != "" {
		if secs, err := strconv.ParseFloat(v, 64); err == nil && secs > 0 {
			return time.Duration(secs * float64(time.Second))
		}
	}
	var body struct {
		RetryAfter float64 `json:"retry_after"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err == nil && body.RetryAfter > 0 {
		return time.Duration(body.RetryAfter * float64(time.Second))
	}
	return discordMinInterval
}

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

// itemDate resolves a feed item's date, preferring gofeed's already-parsed
// values (which handle GMT, MST and many other layouts) and falling back to the
// raw string with our custom layouts. Returns the zero time when unknown.
func itemDate(item *gofeed.Item, layouts []string) time.Time {
	if item.PublishedParsed != nil {
		return *item.PublishedParsed
	}
	if item.UpdatedParsed != nil {
		return *item.UpdatedParsed
	}

	raw := strings.TrimSpace(item.Published)
	if raw == "" {
		raw = strings.TrimSpace(item.Updated)
	}
	if raw == "" {
		return time.Time{}
	}

	parsed, err := parseDate(raw, layouts)
	if err != nil {
		log.Printf("Error parsing published date %q: %v\n", raw, err)
		return time.Time{}
	}
	return parsed
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
			published := itemDate(item, layouts)
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

			// Only notify for genuinely recent items so a one-off backfill of old
			// feed entries does not flood Discord.
			if !published.IsZero() && time.Since(published) <= notifyMaxAge {
				notifyDiscord(newArticle)
			}
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
