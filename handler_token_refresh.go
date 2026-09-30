package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/charlesmariga/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerTokenRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	authHeader := r.Header.Get("Authorization")
	refreshToken, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || strings.TrimSpace(refreshToken) == "" {
		respondWithError(w, http.StatusBadRequest, "Missing or malformed authorization header", errors.New("Bad request"))
		return
	}

	dbRefreshToken, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	if dbRefreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is revoked", nil)
		return
	}

	if time.Now().After(dbRefreshToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is expired", nil)
		return
	}

	token, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate a jwttoken", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{Token: token})
}
