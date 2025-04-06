package service

import (
	"encoding/json"
	"gophre/cmd/data"
	"gophre/env"
	"log"
	"os"
)

func Reset() {
	// Reset articles file
	var allArticles []data.Article
	updatedFile, err := json.MarshalIndent(allArticles, "", "  ")
	if err != nil {
		log.Printf("Error encoding articles JSON: %v\n", err)
		return
	}

	err = os.WriteFile(env.POSTS, updatedFile, 0644)
	if err != nil {
		log.Printf("Error writing articles file: %v\n", err)
		return
	}

	// Reset votes file
	var allVotes []data.Vote
	votesFile, err := json.MarshalIndent(allVotes, "", "  ")
	if err != nil {
		log.Printf("Error encoding votes JSON: %v\n", err)
		return
	}

	err = os.WriteFile(env.VOTES, votesFile, 0644)
	if err != nil {
		log.Printf("Error writing votes file: %v\n", err)
		return
	}
}

