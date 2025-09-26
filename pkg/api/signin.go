package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
)

var password string

func SetPassword() {
	password = os.Getenv("TODO_PASSWORD")
}

func createToken(s string) string {
	secret := []byte("The-secret-word-must-be-replaced") //заменить при сборке
	hashPassword := sha256.Sum256([]byte(s))
	claims := jwt.MapClaims{
		"password": hashPassword,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		return ""
	}
	return signedToken
}

func auth(next http.HandlerFunc, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if password != "" {
			var jwt string

			cookie, err := req.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}

			if createToken(password) != jwt {
				logger.Printf("WARN: Authentification required")
				writeJsonError(res, http.StatusUnauthorized, "Authentification required")
				return
			}
		}
		next(res, req)
	})
}

func signin(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodPost {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var buf bytes.Buffer
	var userPassword struct {
		Password string `json:"password"`
	}

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		logger.Printf("WARN: request reading error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &userPassword); err != nil {
		logger.Printf("WARN: unmarshal error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}

	if userPassword.Password != password {
		logger.Printf("WARN: uncorrect password")
		writeJsonError(res, http.StatusMethodNotAllowed, "Uncorrect password")
		return
	}

	token := createToken(password)
	if token == "" {
		logger.Printf("WARN: creating token error")
		writeJsonError(res, http.StatusMethodNotAllowed, "Creating token error")
		return
	}

	logger.Printf("INFO: user authentication was successful. The token is generated.")
	writeJson(res, http.StatusOK, map[string]any{"token": token})
}
