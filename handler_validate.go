package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
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

	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: strings.Join(words, " "),
	})
}
