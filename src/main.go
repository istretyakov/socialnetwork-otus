package main

import (
	"cloud.google.com/go/civil"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"time"
)

var jwtKey = []byte("jeWA()RU#HHfwajildjwadirh@Hbndwaoj")

type Claims struct {
	Id uuid.UUID `json:"id"`
	jwt.RegisteredClaims
}

type User struct {
	Id         uuid.UUID
	FirstName  string
	SecondName string
	BirthDate  civil.Date
	Biography  string
	City       string
	Password   string
}

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

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	viper.BindEnv("database.hostname", "DATABASE_HOSTNAME")

	Hostname = viper.GetString("database.hostname")

	router := mux.NewRouter()

	router.HandleFunc("/login", LoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/user/register", RegisterUserHandler).Methods(http.MethodPost)
	router.HandleFunc("/user/get/{id}", TokenAuthMiddleware(GetUserByIdHandler)).Methods(http.MethodGet)

	err := http.ListenAndServe(":8080", router)

	if err != nil {
		log.Fatalf("Error starting server, %s", err)
	}
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
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetUserByIdDtoOut{
		FirstName:  user.FirstName,
		SecondName: user.SecondName,
		BirthDate:  user.BirthDate,
		Biography:  user.Biography,
		City:       user.City,
	})
}

func TokenAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
