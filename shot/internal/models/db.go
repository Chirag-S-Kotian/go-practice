package models

import (
	"database/sql"
	"log"
	"os"
	"shot/internal/utils"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	utils.LoadEnv()
	connStr := utils.GetEnv("DB_URL")

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Error connecting to DB:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("❌ DB unreachable:", err)
	}

	log.Println("✅ Connected to Neon ..")

	runMigrations()
}

func runMigrations() {
	schema, err := os.ReadFile("internal/migrations/schema.sql")
	if err != nil {
		log.Fatalf("❌ Failed to read schema.sql: %v", err)
	}

	_, err = DB.Exec(string(schema))
	if err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	log.Println("✅ Database schema migrated")
}