package web

import (
	"encoding/json"
	"gophre/env"
	"gophre/pkg/rss"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"golang.org/x/time/rate"
)

// Gin web framework setuo
// Custom functions
// Rate limiter
// Routes

type RSSFeed struct {
	Path string `json:"path"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func Serve(port ...int) {

	// Server Port
	goodPort := env.PORT
	if len(port) > 0 {
		goodPort = port[0]
	}

	// Define custom functions
	customFuncs := template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"tojson": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a) // on injecte direct, donc pas d'échappement HTML
		},
	}

	// Rate limiter
	limiter := NewLimiter(rate.Every(1*time.Second), 500) // Adjust these values as needed

	// Gin setup
	r := gin.Default()
	// Trust the proxy
	r.SetTrustedProxies([]string{"169.155.237.69"})
	r.SetFuncMap(customFuncs)
	r.Use(RateLimiterMiddleware(limiter))

	// Set up sessions
	store := cookie.NewStore([]byte("secret-session-key")) // Change this to a secure key
	r.Use(sessions.Sessions("gophre-session", store))

	// Configure Goth
	gothic.Store = store

	// Setup auth routes and middleware
	SetupAuth(r)

	// Admin routes
	adminGroup := r.Group("/admin")
	adminGroup.Use(RequireAuth())  // Apply RequireAuth middleware
	adminGroup.Use(RequireAdmin()) // Apply RequireAdmin middleware (placeholder)
	{
		adminGroup.GET("/rss", GetRssFeedsEditorHandler)
		adminGroup.POST("/rss", PostRssFeedsEditorHandler)
	}

	r.LoadHTMLGlob(env.PATH + "assets/html/*")
	r.Static("/css", env.PATH+"assets/css")
	r.Static("/gfx", env.PATH+"assets/gfx")
	r.Static("/js", env.PATH+"assets/js")

	// Website
	r.GET("/", func(c *gin.Context) {
		user, _ := GetCurrentUser(c)
		c.HTML(http.StatusOK, "wall.html", gin.H{
			"user": user,
		})
	})

	r.GET("/all", func(c *gin.Context) {
		user, _ := GetCurrentUser(c)
		c.HTML(http.StatusOK, "wall.html", gin.H{
			"all":  true,
			"user": user,
		})
	})

	r.GET("/me", UserBookmarks)

	r.GET("/top", func(c *gin.Context) {
		user, _ := GetCurrentUser(c)
		c.HTML(http.StatusOK, "wall.html", gin.H{
			"top":  true,
			"user": user,
		})
	})

	// API POST
	r.POST("/vote", func(c *gin.Context) {
		url := c.Query("url")
		note := c.Query("note")
		// Find the article by URL and update its vote
		rss.UpdateVoteByURL(url, note)
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// New vote endpoint using ID
	r.POST("/vote/:id/:vote", Vote)

	// API GET
	r.GET("/article/:id", Article)
	r.GET("/feed", Feed)
	r.GET("/search", func(c *gin.Context) {
		q := c.Query("q")
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}
		size, err := strconv.Atoi(c.DefaultQuery("size", "30"))
		if err != nil || size < 1 || size > 100 {
			size = 30
		}

		// Get all articles but limit the batch size to avoid loading everything
		articles := rss.Feed("", "1", "1000")
		searchResults := rss.Search(q, articles, page, size)
		c.JSON(200, searchResults)
	})

	// Check all articles from one topic
	r.GET("/:path", Topic)

	// Start the server
	r.Run(":" + strconv.Itoa(goodPort))

}

// It returns the article by ID
func Article(c *gin.Context) {
	uid := c.Param("id")
	c.JSON(200, rss.Article(uid))
}

// It takes a path, page and number of elements per page (size), and returns a JSON response with the RSS feed
func Feed(c *gin.Context) {
	path := c.DefaultQuery("path", "")
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "30")
	c.JSON(200, rss.Feed(path, page, size))
}

// It takes the path parameter from the URL, and passes it to the Topic function in the rss package.
// The Topic function returns a slice of articles, which is then passed to the search.html template
func Topic(c *gin.Context) {
	path := c.Param("path")
	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "wall.html", gin.H{
		"path": path,
		"user": user,
	})
}

// GetRssFeedsEditorHandler handles GET requests to /admin/rss
func GetRssFeedsEditorHandler(c *gin.Context) {
	filePath := env.PATH + "rss/rss_feeds.json"
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading RSS feeds file")
		return
	}

	var feeds []RSSFeed
	json.Unmarshal(fileData, &feeds)

	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "admin-rss-editor.html", gin.H{
		"FeedsContent": feeds,
		"user":         user,
	})
}

// PostRssFeedsEditorHandler handles POST requests to /admin/rss
func PostRssFeedsEditorHandler(c *gin.Context) {
	feedsContent := c.PostForm("feeds_content")
	filePath := env.PATH + "rss/rss_feeds.json"

	err := os.WriteFile(filePath, []byte(feedsContent), 0644)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error saving RSS feeds file")
		return
	}

	c.Redirect(http.StatusFound, "/admin/rss")
}
