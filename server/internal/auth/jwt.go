package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var signingKey = []byte(os.Getenv("TOKEN_PRIVATE_KEY"))

type validationResult struct {
	UserID      []byte
	Permissions *permissions
}

func createToken(user userEntry) string {
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
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		panic(err)
	}
	return tokenString
}

func validateToken(tokenString string) (validationResult, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})
	if err != nil {
		return validationResult{}, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return validationResult{}, errors.New("failed to cast token claims")
	}
	userIDString, ok := claims["userId"].(string)
	if !ok {
		return validationResult{}, errors.New("failed to cast user id")
	}
	userID, err := base64.StdEncoding.DecodeString(userIDString)
	if err != nil {
		return validationResult{}, errors.New("failed to decode user id")
	}
	jsonString, _ := json.Marshal(claims["perms"])
	perms := permissions{}
	json.Unmarshal(jsonString, &perms)

	return validationResult{
		UserID:      userID,
		Permissions: &perms,
	}, nil
}
