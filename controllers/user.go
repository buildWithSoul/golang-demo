package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"../models"
	"../repository/user"
	"../utils"

	"github.com/davecgh/go-spew/spew"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func (c Controller) Signup(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		var error models.Error

		json.NewDecoder(r.Body).Decode(&user)

		if user.Email == "" {
			error.Message = "Email is missing."
			utils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		if user.Password == "" {
			// respond with error
			error.Message = "Password is missing."
			// send status bad request
			utils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)

		if err != nil {
			log.Fatal(err)
		}

		user.Password = string(hash)

		spew.Dump(user)
		userRepo := userRepository.UserRepository{}
		user, err = userRepo.Signup(db, user)

		if err != nil {
			error.Message = "Server error"
			utils.RespondWithError(w, http.StatusInternalServerError, error)
			return
		}

		user.Password = ""

		utils.RespondWithJSON(w, user)
	}

}

func (c Controller) Login(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		var jwt models.JWT
		var error models.Error

		// get user data
		json.NewDecoder(r.Body).Decode(&user)

		// is email passed nill?
		if user.Email == "" {
			error.Message = "Email is missing"
			utils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}
		// is password passed nill?
		if user.Password == "" {
			error.Message = "Password is missing"
			utils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		// does user exist?

		//save pwd before query
		password := user.Password

		userRepo := userRepository.UserRepository{}
		user, err := userRepo.Login(db, user)

		if err != nil {
			if err == sql.ErrNoRows {
				error.Message = "The user doesn't exist"
				utils.RespondWithError(w, http.StatusBadRequest, error)
				return
			}
			log.Fatal(err)
		}

		// is token good? hash match
		hashedPassword := user.Password

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

		if err != nil {
			error.Message = "Passwords don't match"
			utils.RespondWithError(w, http.StatusBadRequest, error)
			return
		}

		// passes crypto, start token gen
		token, err := utils.GenerateToken(user)

		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		jwt.Token = token

		utils.RespondWithJSON(w, jwt)
	}
}

func (c Controller) TokenVerifyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error
		authHeader := r.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) == 2 {
			authToken := bearerToken[1]
			token, error := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}

				return []byte(os.Getenv("SECRET")), nil
			})

			if error != nil {
				errorObject.Message = error.Error()
				utils.RespondWithError(w, http.StatusUnauthorized, errorObject)
			}

			if token.Valid {
				next.ServeHTTP(w, r)
			}

		} else {
			errorObject.Message = "Not JWT included"
			utils.RespondWithError(w, http.StatusBadRequest, errorObject)
		}

	})
}
