package main

import (
	"net/http"
	"slices"

	"github.com/charlesmariga/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "" {
		sortOrder = "asc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		respondWithError(w, http.StatusBadRequest, "sort must be 'asc' or 'desc'", nil)
		return
	}

	var dbChirps []database.Chirp
	var err error

	s := r.URL.Query().Get("author_id")
	if s != "" {
		authorID, parseErr := uuid.Parse(s)
		if parseErr != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author_id", parseErr)
			return
		}
		dbChirps, err = cfg.db.GetAllChirpsByUserID(r.Context(), authorID)
	} else {
		dbChirps, err = cfg.db.GetAllChirps(r.Context())
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserId:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}

	// The queries already return rows in ascending created_at order.
	if sortOrder == "desc" {
		slices.SortFunc(chirps, func(a, b Chirp) int {
			return b.CreatedAt.Compare(a.CreatedAt)
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}
