package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/charlesmariga/chirpy/internal/auth"
	"github.com/charlesmariga/chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	if !match {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	hourInSeconds := 60 * 60 * time.Second
	var expiresIn time.Duration
	if params.ExpiresInSeconds > 0 && time.Duration(params.ExpiresInSeconds) < hourInSeconds {
		expiresIn = time.Duration(params.ExpiresInSeconds)
	} else {
		expiresIn = hourInSeconds
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate a jwt token", err)
		return
	}

	sixtyDaysInSeconds := 60 * 24 * 60 * 60 * time.Second
	refreshToken := auth.MakeRefreshToken()
	dbRefreshToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().Add(sixtyDaysInSeconds),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couln't generate a refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
		},
		Token:        token,
		RefreshToken: dbRefreshToken.Token,
	})
}
