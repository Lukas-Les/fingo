package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Lukas-Les/fingo/internal/auth"
	"github.com/Lukas-Les/fingo/internal/database"
	"github.com/google/uuid"
)

type userQueries interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
}

func BuildUserCreateHandler(db userQueries) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO: move this to a validateCreds middleware
		type parameters struct {
			Password string `json:"password"`
			Email    string `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
			return
		}

		if params.Password == "" || params.Email == "" {
			respondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
			return
		}

		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
			return
		}
		// TODO: move this to a validateCreds middleware

		user, err := db.CreateUser(r.Context(), database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
			return
		}

		respondWithJSON(w, http.StatusCreated, user)
	}
}

func BuildUserLoginHandler(db userQueries, jwtSecret string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			ID           uuid.UUID `json:"id"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			Email        string    `json:"email"`
			Token        string    `json:"token"`
			RefreshToken string    `json:"refresh_token"`
		}

		// TODO: move this to a validateCreds middleware
		type parameters struct {
			Password string `json:"password"`
			Email    string `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
			return
		}

		if params.Password == "" || params.Email == "" {
			respondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
			return
		}
		// TODO: move this to a validateCreds middleware

		// TODO: move this to a validatePassword milldeware
		user, err := db.GetUserByEmail(r.Context(), params.Email)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error(), err)
			return
		}
		isValid, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error(), err)
			return
		}
		if !isValid {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}
		// TODO: move this to a validatePassword milldeware

		jwtToken, err := auth.MakeJWT(user.ID, jwtSecret, time.Minute)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error(), err)
			return
		}

		resp := response{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        jwtToken,
			RefreshToken: "",
		}
		respondWithJSON(w, http.StatusOK, resp)
	}
}
