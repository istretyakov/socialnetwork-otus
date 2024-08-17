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

type User struct {
	Id         uuid.UUID
	FirstName  string
	SecondName string
	BirthDate  civil.Date
	Biography  string
	City       string
	Password   string
}

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
		return errors.New("error connecting to database")
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
		return User{}, errors.New("error connecting to database")
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

func SearchUsersByName(firstname string, lastname string) ([]User, error) {
	db := ConnectPostgres()

	if db == nil {
		return []User{}, errors.New("error connecting to database")
	}

	defer db.Close()

	sql := "SELECT id, first_name, second_name, birthdate, biography, city, password FROM users"

	if firstname != "" && lastname != "" {
		sql += " WHERE first_name LIKE $1 AND second_name LIKE $2"
	} else if firstname != "" {
		sql += " WHERE first_name LIKE $1"
	} else if lastname != "" {
		sql += " WHERE second_name LIKE $1"
	}

	sql += " ORDER BY id"

	rows, err := db.Query(sql, firstname, lastname)

	if err != nil {
		return []User{}, err
	}

	users := []User{}

	for rows.Next() {
		user := User{}
		var birthDate time.Time

		err := rows.Scan(&user.Id, &user.FirstName, &user.SecondName, &birthDate, &user.Biography, &user.City, &user.Password)

		if err != nil {
			return []User{}, err
		}

		user.BirthDate = civil.DateOf(birthDate)

		users = append(users, user)
	}

	return users, nil
}
