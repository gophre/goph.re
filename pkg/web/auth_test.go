package web

import (
	"encoding/json"
	"gophre/cmd/data"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Global session store for tests, initialized once.
var testStore sessions.Store

func init() {
	gin.SetMode(gin.TestMode)
	// Initialize the session store, using the same key as in the application
	testStore = cookie.NewStore([]byte("secret-session-key"))
}

// TestRequireAdmin_UnauthenticatedUser tests the RequireAdmin middleware when the user is not logged in.
// Expected: Redirect to homepage (HTTP 302 to "/").
func TestRequireAdmin_UnauthenticatedUser(t *testing.T) {
	// Setup: Create a Gin engine and response recorder.
	r := gin.New()
	// Apply session middleware
	r.Use(sessions.Sessions("gophre-session", testStore))
	// Apply RequireAdmin middleware
	r.Use(RequireAdmin())
	// Dummy handler that should not be reached
	r.GET("/admin/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/test", nil)

	// Execution: Call the RequireAdmin() middleware by serving the request.
	r.ServeHTTP(w, req)

	// Assertion: Check response code is 302 and Location header is "/".
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))
}

// TestRequireAdmin_AuthenticatedNonAdminUser tests the RequireAdmin middleware
// when a non-admin user is logged in.
// Expected: Redirect to homepage (HTTP 302 to "/").
func TestRequireAdmin_AuthenticatedNonAdminUser(t *testing.T) {
	// Setup: Create a Gin engine and response recorder.
	r := gin.New()
	// Apply session middleware
	r.Use(sessions.Sessions("gophre-session", testStore))

	// Middleware to set up a non-admin user session for this test
	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		user := data.User{UserID: "user123", Name: "nonAdminUser", Email: "test@example.com", Provider: "github"}
		userDataJSON, err := json.Marshal(user)
		assert.NoError(t, err)
		session.Set("user", string(userDataJSON))
		err = session.Save()
		assert.NoError(t, err)
		c.Next()
	})

	// Apply RequireAdmin middleware
	r.Use(RequireAdmin())
	// Dummy handler that should not be reached
	r.GET("/admin/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	// Perform a request to a route that will trigger the session setup and then RequireAdmin
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	
	// Execution: Call the RequireAdmin() middleware by serving the request.
	r.ServeHTTP(w, req)

	// Assertion: Check response code is 302 and Location header is "/".
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))
}

// TestRequireAdmin_AuthenticatedAdminUser tests the RequireAdmin middleware
// when the admin user 'l3dlp' is logged in.
// Expected: Request proceeds (HTTP 200 from dummy handler).
func TestRequireAdmin_AuthenticatedAdminUser(t *testing.T) {
	// Setup: Create a Gin engine and response recorder.
	r := gin.New()
	// Use sessions.Sessions on the engine
	r.Use(sessions.Sessions("gophre-session", testStore))

	// Middleware to set up an admin user session for this test
	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		adminUser := data.User{UserID: "admin456", Name: "l3dlp", Email: "admin@example.com", Provider: "github"}
		userDataJSON, err := json.Marshal(adminUser)
		assert.NoError(t, err)
		session.Set("user", string(userDataJSON))
		err = session.Save()
		assert.NoError(t, err)
		c.Next()
	})

	// Apply the RequireAdmin() middleware.
	r.Use(RequireAdmin())
	// Add a dummy handler after the middleware that sets a 200 OK status.
	r.GET("/admin/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Admin access granted")
	})

	w := httptest.NewRecorder()
	// Perform a request to a test route.
	req, _ := http.NewRequest("GET", "/admin/test", nil)

	// Execution: Call the RequireAdmin() middleware by serving the request.
	r.ServeHTTP(w, req)

	// Assertion: Check response code is 200.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Admin access granted", w.Body.String())
}
