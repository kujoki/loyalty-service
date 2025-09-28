package handler

import (
    "time"
	"errors"
	"log"
    "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("my_secret_key") // внести в виртуальное окружение

var ErrNoUserLogin = errors.New("there is no user Login")

type Claims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

func GenerateJWT(login string) (string, error) {
    claims := jwt.MapClaims{
        "sub": login,                   
        "iat": time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func ParseUserLogin(tokenString string) (string, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
    return []byte(jwtSecret), nil
	})
	if err != nil {
		log.Println("error during parse", err)
        return "", err
    }

    if !token.Valid {
        log.Println("token is invalid")
        return "", errors.New("there is invalid token")
    }

    if claims.Login == "" {
        log.Println("token does not contain user login")
        return "", ErrNoUserLogin
    }

    log.Println("Token is valid")
    return claims.Login, nil
}