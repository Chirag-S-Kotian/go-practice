package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading env variables")
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
