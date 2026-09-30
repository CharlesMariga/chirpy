package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/charlesmariga/chirpy/internal/auth"
	"github.com/charlesmariga/chirpy/internal/database"
)

func (apiCfg *apiConfig) handleChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type chirp struct {
		Chirp
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	userID, err := auth.ValidateJWT(token, apiCfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	params := parameters{}
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	profane := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(params.Body, " ")

	for _, p := range profane {
		for index, word := range words {
			if strings.ToLower(word) == p {
				words[index] = "****"
			}
		}
	}

	body := strings.Join(words, " ")
	newChirp, err := apiCfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: userID,
		Body:   body,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp{
		Chirp: Chirp{
			ID:        newChirp.ID,
			UserId:    newChirp.UserID,
			CreatedAt: newChirp.CreatedAt,
			UpdatedAt: newChirp.UpdatedAt,
			Body:      newChirp.Body,
		},
	})
}
