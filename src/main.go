package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	environmentName := os.Getenv("ENVIRONMENT")

	if err := initializeConfig(environmentName); err != nil {
		log.Fatalf("Error reading config file, %s", err)
		return
	}

	if err := viper.Unmarshal(&DatabaseSettings); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
		return
	}

	if err := initializeDatabases(DatabaseSettings); err != nil {
		log.Fatalf("Error connecting to database, %s", err)
		return
	}

	for _, readerDbPool := range ReaderDbPools {
		defer readerDbPool.Close()
	}

	router := mux.NewRouter()

	router.HandleFunc("/login", LoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/user/register", RegisterUserHandler).Methods(http.MethodPost)
	router.HandleFunc("/user/get/{id}", TokenAuthMiddleware(GetUserByIdHandler)).Methods(http.MethodGet)
	router.HandleFunc("/user/search", SearchUsersHandler).Methods(http.MethodGet)

	if err := http.ListenAndServe(":8100", router); err != nil {
		log.Fatalf("Error starting server, %s", err)
		return
	}
}

func initializeConfig(environmentName string) error {
	viper.SetConfigName(fmt.Sprint("config.", environmentName))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	err := viper.ReadInConfig()

	return err
}

func initializeDatabases(config Config) error {
	masterDbPool, err := ConnectPostgresUsingPool(DatabaseSettings.Databases.Master)

	if err != nil {
		return err
	}

	MasterDbPool = masterDbPool

	ReaderDbPools = append(ReaderDbPools, MasterDbPool)

	for _, slaveSettings := range DatabaseSettings.Databases.Slaves {
		slaveDbPool, err := ConnectPostgresUsingPool(slaveSettings)

		if err != nil {
			return err
		}

		ReaderDbPools = append(ReaderDbPools, slaveDbPool)
	}

	return nil
}
