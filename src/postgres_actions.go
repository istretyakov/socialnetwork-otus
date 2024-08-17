package main

import (
	"cloud.google.com/go/civil"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
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

var DbPool *pgxpool.Pool

type User struct {
	Id         uuid.UUID
	FirstName  string
	SecondName string
	BirthDate  civil.Date
	Biography  string
	City       string
	Password   string
}

func ConnectPostgresUsingPool() *pgxpool.Pool {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Hostname, Port, Username, Password, Database)

	config, err := pgxpool.ParseConfig(connString)

	if err != nil {
		log.Fatal(err)
	}

	dbPool, err := pgxpool.ConnectConfig(context.Background(), config)

	if err != nil {
		log.Fatal(err)
	}

	return dbPool
}

func InsertUser(user User) error {
	_, err := DbPool.Exec(context.Background(), "INSERT INTO users(id, first_name, second_name, birthdate, biography, city, password) VALUES($1, $2, $3, $4, $5, $6, $7)",
		user.Id.String(), user.FirstName, user.SecondName, user.BirthDate.String(), user.Biography, user.City, user.Password)

	if err != nil {
		return err
	}

	return err
}

func FindUserById(id uuid.UUID) (User, error) {
	row := DbPool.QueryRow(context.Background(), "SELECT id, first_name, second_name, birthdate, biography, city, password FROM users WHERE id = $1", id)

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
	sql := "SELECT id, first_name, second_name, birthdate, biography, city, password FROM users"

	if firstname != "" && lastname != "" {
		sql += " WHERE first_name LIKE $1 AND second_name LIKE $2"
	} else if firstname != "" {
		sql += " WHERE first_name LIKE $1"
	} else if lastname != "" {
		sql += " WHERE second_name LIKE $1"
	}

	sql += " ORDER BY id"

	rows, err := DbPool.Query(context.Background(), sql, firstname, lastname)

	if err != nil {
		return []User{}, err
	}

	defer rows.Close()

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
