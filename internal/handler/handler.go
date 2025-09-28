package handler

import (
	"errors"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"log"
	"encoding/json"
	"github.com/kujoki/loyalty-service/internal/repository"
	"github.com/kujoki/loyalty-service/internal/model"
)

type AuthRequest struct {
	Login	    string  `json:"login"`
	Password    string  `json:"password"`
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func SetAuthCookie(login string, w http.ResponseWriter) error {
	token, err := GenerateJWT(login)
	if err != nil {
		log.Println("error jwt generate:", err)
		return err
	}

	http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
		})

	w.Header().Set("Authorization", "Bearer "+token)

	return nil
}

func WrapperHandlerAuth(r *repository.PostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var jsonReq AuthRequest
		if err := json.NewDecoder(req.Body).Decode(&jsonReq); err != nil {
			log.Println("error decoding request:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hashedPwd, err := HashPassword(jsonReq.Password)
		if err != nil {
			log.Println("error hashing password:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if er := r.SetAuth(req.Context(), jsonReq.Login, hashedPwd); er != nil {
			if errors.Is(er, model.ErrLoginExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			log.Println("error saving auth:", er)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := SetAuthCookie(jsonReq.Login, w); err != nil {
			log.Println("error setting cookie:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func WrapperHandlerLogin(r *repository.PostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var jsonReq AuthRequest
		if err := json.NewDecoder(req.Body).Decode(&jsonReq); err != nil {
			log.Println("error decoding request:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		storedHash, err := r.GetHashPassword(req.Context(), jsonReq.Login)
		if err != nil {
			log.Println("user not found:", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !CheckPasswordHash(jsonReq.Password, storedHash) {
			log.Println("invalid password for login:", jsonReq.Login)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := SetAuthCookie(jsonReq.Login, w); err != nil {
			log.Println("error setting cookie:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}