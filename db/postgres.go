package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// var DB *sql.DB

// var Queries *db.Queries

func NewDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Cannot reach database:", err)
	}
	fmt.Println("Connected to database!")

	return db
}
