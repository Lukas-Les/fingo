package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type credentials struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (c credentials) validate() error {
	if c.Password == "" || c.Email == "" {
		return errors.New("missing required fields")
	}
	return nil
}

func credentialsFromRequest(r *http.Request) (credentials, error) {
	var c credentials
	if err := decodeJsonRequest(&c, r); err != nil {
		return c, fmt.Errorf("decode credentials: %w", err)
	}
	if err := c.validate(); err != nil {
		return c, err
	}
	return c, nil
}

func decodeJsonRequest[T any](v *T, r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

func BuildUserCreateHandler(db userQueries) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO: move this to a validateCreds middleware
		creds, err := credentialsFromRequest(r)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error(), err)
			return
		}
		hashedPassword, err := auth.HashPassword(creds.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
			return
		}
		// TODO: move this to a validateCreds middleware

		user, err := db.CreateUser(r.Context(), database.CreateUserParams{
			Email:          creds.Email,
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
		creds, err := credentialsFromRequest(r)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error(), err)
			return
		}
		// TODO: move this to a validateCreds middleware

		// TODO: move this to a validatePassword milldeware
		user, err := db.GetUserByEmail(r.Context(), creds.Email)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error(), err)
			return
		}
		isValid, err := auth.CheckPasswordHash(creds.Password, user.HashedPassword)
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
