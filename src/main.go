package main

import (
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
		return
	}

	viper.BindEnv("database.hostname", "DATABASE_HOSTNAME")

	Hostname = viper.GetString("database.hostname")

	router := mux.NewRouter()

	router.HandleFunc("/login", LoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/user/register", RegisterUserHandler).Methods(http.MethodPost)
	router.HandleFunc("/user/get/{id}", TokenAuthMiddleware(GetUserByIdHandler)).Methods(http.MethodGet)
	router.HandleFunc("/user/search", SearchUsersHandler).Methods(http.MethodGet)

	DbPool = ConnectPostgresUsingPool()
	defer DbPool.Close()

	err := http.ListenAndServe(":8080", router)

	if err != nil {
		log.Fatalf("Error starting server, %s", err)
	}
}
