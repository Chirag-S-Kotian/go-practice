package models

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"shot/internal/utils"
)

var DB *sql.DB

func InitDB() {
	utils.LoadEnv()
	connStr := utils.GetEnv("DB_URL")

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to DB:", err)
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal("db unreachable:", err)
	}
	log.Println("connected to neon ..")
}
