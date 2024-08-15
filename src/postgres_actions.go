package main

import (
	"cloud.google.com/go/civil"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"time"
)

var (
	Hostname = "socialnetwork_postgres"
	Port     = "5432"
	Username = "postgres"
	Password = "Test1234"
	Database = "social_network"
)

func ConnectPostgres() *sql.DB {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Hostname, Port, Username, Password, Database)

	db, err := sql.Open("postgres", connString)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return db
}

func InsertUser(user User) error {
	db := ConnectPostgres()

	if db == nil {
		return errors.New("Error connecting to database")
	}

	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO users(id, first_name, second_name, birthdate, biography, city, password) VALUES($1, $2, $3, $4, $5, $6, $7)")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(user.Id.String(), user.FirstName, user.SecondName, user.BirthDate.String(), user.Biography, user.City, user.Password)

	return err
}

func FindUserById(id uuid.UUID) (User, error) {
	db := ConnectPostgres()

	if db == nil {
		return User{}, errors.New("Error connecting to database")
	}

	defer db.Close()

	row := db.QueryRow("SELECT id, first_name, second_name, birthdate, biography, city, password FROM users WHERE id = $1", id)

	user := User{}

	var birthDate time.Time

	err := row.Scan(&user.Id, &user.FirstName, &user.SecondName, &birthDate, &user.Biography, &user.City, &user.Password)

	user.BirthDate = civil.DateOf(birthDate)

	if err != nil {
		return User{}, err
	}

	return user, nil
}
