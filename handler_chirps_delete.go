package main

import (
	"net/http"

	"github.com/charlesmariga/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
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

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID", err)
		return
	}

	dbChirp, err := apiCfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden", err)
		return
	}

	err = apiCfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
