package main

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

func (cfg *apiConfig) handlerRefreshTokenRevoke(w http.ResponseWriter, r *http.Request) {
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
		respondWithError(w, http.StatusUnauthorized, "Refresh token is already revoked", nil)
		return
	}

	if time.Now().After(dbRefreshToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is expired", nil)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), dbRefreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token was not revoked", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
