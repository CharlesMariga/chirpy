package main

import (
	"encoding/json"
	"net/http"

	"github.com/charlesmariga/chirpy/internal/auth"
	"github.com/charlesmariga/chirpy/internal/database"
)

func (apiCfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
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

	params := &parameter{}
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	dbUser, err := apiCfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        dbUser.ID,
			Email:     dbUser.Email,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		},
	})
}
