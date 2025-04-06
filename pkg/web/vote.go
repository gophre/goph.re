package web

import (
	"encoding/json"
	"gophre/cmd/data"
	"gophre/env"
	"gophre/pkg/rss"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func Vote(c *gin.Context) {
	id := c.Param("id")
	voteValue := c.Param("vote")
	
	// Récupérer l'utilisateur courant
	user, isAuthenticated := GetCurrentUser(c)
	if !isAuthenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Authentication required"})
		return
	}
	
	// Créer un nouveau vote
	newVote := data.Vote{
		ArticleID: id,
		UserID:    user.ID,
		Timestamp: time.Now(),
	}
	
	if voteValue == "GOOD" {
		newVote.Value = 1
	} else if voteValue == "BAD" {
		newVote.Value = -1
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid vote value"})
		return
	}
	
	// Lire les votes actuels
	var votes []data.Vote
	file, err := os.ReadFile(env.VOTES)
	if err == nil {
		err = json.Unmarshal(file, &votes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to parse votes"})
			return
		}
	}
	
	// Vérifier si l'utilisateur a déjà voté pour cet article
	existingVoteIndex := -1
	for i, vote := range votes {
		if vote.ArticleID == id && vote.UserID == user.ID {
			existingVoteIndex = i
			break
		}
	}
	
	// Soit mettre à jour le vote existant, soit ajouter un nouveau
	if existingVoteIndex >= 0 {
		// Mettre à jour le vote existant
		votes[existingVoteIndex] = newVote
	} else {
		// Ajouter un nouveau vote
		votes = append(votes, newVote)
	}
	
	// Sauvegarder les votes
	updatedFile, err := json.MarshalIndent(votes, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to encode votes"})
		return
	}
	
	err = os.WriteFile(env.VOTES, updatedFile, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save votes"})
		return
	}
	
	// Mettre à jour le fichier de préférences de l'utilisateur
	updateUserPreferences(user.ID, id, voteValue)
	
	// Mettre également à jour le vote de l'article dans le fichier posts
	// Cela maintient la compatibilité avec le code existant
	updateArticleVoteAggregate(id)
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Mettre à jour les préférences utilisateur (remplace le localStorage)
func updateUserPreferences(userID, articleID, voteValue string) error {
	userFilePath := filepath.Join(env.USERS, userID+".json")
	
	var preferences map[string]interface{}
	
	// Lire les préférences existantes
	file, err := os.ReadFile(userFilePath)
	if err != nil {
		// Si le fichier n'existe pas, créer un nouveau
		preferences = map[string]interface{}{
			"user_id": userID,
			"votes":   map[string]string{},
		}
	} else {
		// Sinon, décoder les préférences existantes
		err = json.Unmarshal(file, &preferences)
		if err != nil {
			return err
		}
	}
	
	// Vérifier si la clé de votes existe, sinon la créer
	votes, ok := preferences["votes"].(map[string]interface{})
	if !ok {
		votes = make(map[string]interface{})
		preferences["votes"] = votes
	}
	
	// Mettre à jour la valeur du vote
	voteKey := "vote_" + articleID
	votes[voteKey] = voteValue
	
	// Sauvegarder le fichier mis à jour
	updatedFile, err := json.MarshalIndent(preferences, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(userFilePath, updatedFile, 0644)
}

// UpdateArticleVoteAggregate met à jour le champ vote de l'article avec la valeur de vote agrégée
func updateArticleVoteAggregate(id string) error {
	// Obtenir le nombre total de votes
	totalVotes := GetArticleVotesTotal(id)
	
	// Lire les articles actuels depuis le fichier
	file, err := os.ReadFile(env.POSTS)
	if err != nil {
		return err
	}

	var articles []data.Article
	err = json.Unmarshal(file, &articles)
	if err != nil {
		return err
	}

	// Mettre à jour le vote pour l'article avec l'ID correspondant
	articleFound := false
	for i := range articles {
		if articles[i].ID == id {
			articleFound = true
			articles[i].Vote = totalVotes
			break
		}
	}
	
	if !articleFound {
		return nil
	}

	// Écrire les articles mis à jour dans le fichier
	updatedFile, err := json.MarshalIndent(articles, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(env.POSTS, updatedFile, 0644)
}

// GetArticleVotesTotal renvoie le total de tous les votes pour un article
func GetArticleVotesTotal(articleID string) int {
	var votes []data.Vote
	file, err := os.ReadFile(env.VOTES)
	if err != nil {
		// Si le fichier de votes n'existe pas encore, revenir au champ vote de l'article
		return rss.Article(articleID).Vote
	}
	
	err = json.Unmarshal(file, &votes)
	if err != nil {
		return 0
	}
	
	total := 0
	for _, vote := range votes {
		if vote.ArticleID == articleID {
			total += vote.Value
		}
	}
	
	return total
}

// GetUserVote vérifie si un utilisateur a déjà voté pour un article
func GetUserVote(articleID, userID string) (bool, int) {
	var votes []data.Vote
	file, err := os.ReadFile(env.VOTES)
	if err != nil {
		return false, 0
	}
	
	err = json.Unmarshal(file, &votes)
	if err != nil {
		return false, 0
	}
	
	for _, vote := range votes {
		if vote.ArticleID == articleID && vote.UserID == userID {
			return true, vote.Value
		}
	}
	
	return false, 0
}