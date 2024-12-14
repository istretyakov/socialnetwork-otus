package main

import (
	"cloud.google.com/go/civil"
	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"time"
)

type LoginDtoIn struct {
	Id       uuid.UUID `json:"id"`
	Password string    `json:"password"`
}

type CreateUserDtoIn struct {
	FirstName  string     `json:"first_name"`
	SecondName string     `json:"second_name"`
	BirthDate  civil.Date `json:"birthdate"`
	Biography  string     `json:"biography"`
	City       string     `json:"city"`
	Password   string     `json:"password"`
}

type GetUserByIdDtoOut struct {
	FirstName  string     `json:"first_name"`
	SecondName string     `json:"second_name"`
	BirthDate  civil.Date `json:"birthdate"`
	Biography  string     `json:"biography"`
	City       string     `json:"city"`
}

type SearchUsersDtoOut struct {
	FirstName  string     `json:"first_name"`
	SecondName string     `json:"second_name"`
	BirthDate  civil.Date `json:"birthdate"`
	Biography  string     `json:"biography"`
	City       string     `json:"city"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if len(d) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var loginDtoIn = LoginDtoIn{}
	err = json.Unmarshal(d, &loginDtoIn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	foundUser, err := FindUserById(loginDtoIn.Id)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(loginDtoIn.Password))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Id: foundUser.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		log.Println("Error signing the token: %v", err)
		http.Error(w, "Error signing the token", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(tokenString))
}

func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if len(d) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var createUserDtoIn = CreateUserDtoIn{}
	err = json.Unmarshal(d, &createUserDtoIn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(createUserDtoIn.Password), 10)

	user := User{
		Id:         uuid.New(),
		FirstName:  createUserDtoIn.FirstName,
		SecondName: createUserDtoIn.SecondName,
		BirthDate:  createUserDtoIn.BirthDate,
		Biography:  createUserDtoIn.Biography,
		City:       createUserDtoIn.City,
		Password:   string(hashPassword),
	}

	err = InsertUser(user)

	if err != nil {
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(user.Id.String()))
}

func GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := FindUserById(id)

	if err != nil {   
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetUserByIdDtoOut{
		FirstName:  user.FirstName,
		SecondName: user.SecondName,
		BirthDate:  user.BirthDate,
		Biography:  user.Biography,
		City:       user.City,
	})
}

func SearchUsersHandler(w http.ResponseWriter, r *http.Request) {
	firstname := r.URL.Query().Get("first_name")
	lastname := r.URL.Query().Get("last_name")

	if len(firstname) == 0 || len(lastname) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Невалидные данные"))
		return
	}

	users, err := SearchUsersByName(firstname, lastname)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	searchUsersDtoOut := make([]SearchUsersDtoOut, len(users))

	for i, user := range users {
		searchUsersDtoOut[i] = SearchUsersDtoOut{
			FirstName:  user.FirstName,
			SecondName: user.SecondName,
			BirthDate:  user.BirthDate,
			Biography:  user.Biography,
			City:       user.City,
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchUsersDtoOut)
}
