package main

import (
	"cloud.google.com/go/civil"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"time"
)

type DatabaseConfig struct {
	Hostname string `mapstructure:"hostname"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type Config struct {
	Databases struct {
		Master DatabaseConfig   `mapstructure:"master"`
		Slaves []DatabaseConfig `mapstructure:"slaves"`
	} `mapstructure:"databases"`
}

var DatabaseSettings Config

var MasterDbPool *pgxpool.Pool
var ReaderDbPools []*pgxpool.Pool

type User struct {
	Id         uuid.UUID
	FirstName  string
	SecondName string
	BirthDate  civil.Date
	Biography  string
	City       string
	Password   string
}

func ConnectPostgresUsingPool(databaseConfig DatabaseConfig) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		databaseConfig.Hostname,
		databaseConfig.Port,
		databaseConfig.Username,
		databaseConfig.Password,
		databaseConfig.Database)

	config, err := pgxpool.ParseConfig(connString)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	dbPool, err := pgxpool.ConnectConfig(context.Background(), config)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return dbPool, nil
}

func GetReaderDbPool() *pgxpool.Pool {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return ReaderDbPools[r.Intn(len(ReaderDbPools))]
}

func InsertUser(user User) error {
	_, err := MasterDbPool.Exec(context.Background(), "INSERT INTO users(id, first_name, second_name, birthdate, biography, city, password) VALUES($1, $2, $3, $4, $5, $6, $7)",
		user.Id.String(), user.FirstName, user.SecondName, user.BirthDate.String(), user.Biography, user.City, user.Password)

	if err != nil {
		return err
	}

	return err
}

func FindUserById(id uuid.UUID) (User, error) {
	row := GetReaderDbPool().QueryRow(context.Background(), "SELECT id, first_name, second_name, birthdate, biography, city, password FROM users WHERE id = $1", id)

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

	rows, err := GetReaderDbPool().Query(context.Background(), sql, firstname, lastname)

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
