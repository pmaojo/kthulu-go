package main

import (
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"web-test/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [command]")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	command := os.Args[1]
	dir := "migrations"

	// Connect to database using configuration
	db, err := goose.OpenDBWithDriver("sqlite3", cfg.Database.URL)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	arguments := []string{}
	if len(os.Args) > 2 {
		arguments = os.Args[2:]
	}

	if err := goose.Run(command, db, dir, arguments...); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
