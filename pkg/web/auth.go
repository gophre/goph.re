package web

import (
	"encoding/json"
	"gophre/cmd/data"
	"gophre/env"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

// SetupAuth configure l'authentification GitHub
func SetupAuth(router *gin.Engine) {
	// Configure les fournisseurs d'authentification
	goth.UseProviders(
		github.New(env.GITHUB_CLIENT_ID, env.GITHUB_CLIENT_SECRET, env.GITHUB_CALLBACK_URL),
	)

	// Routes d'authentification
	auth := router.Group("/auth")
	{
		auth.GET("/github", func(c *gin.Context) {
			// Set the provider to GitHub
			q := c.Request.URL.Query()
			q.Set("provider", "github")
			c.Request.URL.RawQuery = q.Encode()
			// Démarre le processus d'authentification
			gothic.BeginAuthHandler(c.Writer, c.Request)
		})

		auth.GET("/github/callback", func(c *gin.Context) {
			// Set the provider to GitHub
			q := c.Request.URL.Query()
			q.Set("provider", "github")
			c.Request.URL.RawQuery = q.Encode()
			
			// Récupère l'utilisateur après authentification GitHub
			user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Crée un objet User et le stocke en session
			session := sessions.Default(c)
			userData := data.User{
				ID:        user.UserID,
				Name:      user.Name,
				Email:     user.Email,
				AvatarURL: user.AvatarURL,
				Provider:  user.Provider,
			}

			// Sauvegarde les données utilisateur en session
			userDataJSON, _ := json.Marshal(userData)
			session.Set("user", string(userDataJSON))
			session.Save()

			// Créer le fichier des préférences utilisateur s'il n'existe pas
			createUserPreferencesFile(userData.ID)

			// Redirige vers la page d'accueil
			c.Redirect(http.StatusTemporaryRedirect, "/")
		})

		auth.GET("/logout", func(c *gin.Context) {
			session := sessions.Default(c)
			session.Clear()
			session.Save()
			c.Redirect(http.StatusTemporaryRedirect, "/")
		})
	}
}

// GetCurrentUser récupère l'utilisateur courant à partir de la session
func GetCurrentUser(c *gin.Context) (data.User, bool) {
	session := sessions.Default(c)
	userStr := session.Get("user")
	
	if userStr == nil {
		return data.User{}, false
	}

	var user data.User
	err := json.Unmarshal([]byte(userStr.(string)), &user)
	if err != nil {
		return data.User{}, false
	}

	return user, true
}

// RequireAuth middleware pour vérifier si l'utilisateur est authentifié
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, isAuthenticated := GetCurrentUser(c)
		if !isAuthenticated {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}

// createUserPreferencesFile crée un fichier de préférences vide pour un nouvel utilisateur
func createUserPreferencesFile(userID string) error {
	userFilePath := filepath.Join(env.USERS, userID+".json")
	
	// Vérifie si le fichier existe déjà
	if _, err := os.Stat(userFilePath); os.IsNotExist(err) {
		// Crée un fichier vide avec des préférences par défaut
		preferences := map[string]interface{}{
			"user_id": userID,
			"votes":   map[string]string{}, // Equivalent au localStorage
		}
		
		data, err := json.MarshalIndent(preferences, "", "  ")
		if err != nil {
			return err
		}
		
		return os.WriteFile(userFilePath, data, 0644)
	}
	
	return nil
}