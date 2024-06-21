package auth

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"lath/xman/internal/db"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type validationResult struct {
	UserID      string
	Permissions *permissions
}

var tokenSecret []byte

func Init() {
	s, ok := db.FindServerStateXman()
	if !ok || len(s.TokenSecret) == 0 {
		fmt.Println("Generating new token secret")
		tokenSecret = make([]byte, 40)
		_, err := rand.Read(tokenSecret)
		if err != nil {
			panic(err)
		}
		db.UpsertServerStateXmanTokenSecret(tokenSecret)
	} else {
		tokenSecret = s.TokenSecret
	}
}

func createToken(user userEntry) string {
	if len(tokenSecret) == 0 {
		panic("token secret not initialized")
	}
	token_lifespan, err := strconv.Atoi(os.Getenv("TOKEN_DAY_LIFESPAN"))
	if err != nil {
		panic(err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"perms":  user.Permissions,
		"exp":    time.Now().Add(time.Hour * 24 * time.Duration(token_lifespan)).Unix(),
	})
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(tokenSecret)
	if err != nil {
		panic(err)
	}
	return tokenString
}

func validateToken(tokenString string) (validationResult, error) {
	if len(tokenSecret) == 0 {
		panic("token secret not initialized")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tokenSecret, nil
	})
	if err != nil {
		return validationResult{}, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return validationResult{}, errors.New("failed to cast token claims")
	}
	userID, ok := claims["userId"].(string)
	if !ok {
		return validationResult{}, errors.New("failed to cast user id")
	}
	jsonString, _ := json.Marshal(claims["perms"])
	perms := permissions{}
	json.Unmarshal(jsonString, &perms)

	return validationResult{
		UserID:      userID,
		Permissions: &perms,
	}, nil
}
