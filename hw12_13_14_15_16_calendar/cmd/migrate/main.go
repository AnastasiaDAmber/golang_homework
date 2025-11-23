package main

import (
	"flag"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var (
	dsn        string
	migrations string
)

func init() {
	flag.StringVar(&dsn, "dsn", "", "PostgreSQL DSN")
	flag.StringVar(&migrations, "migrations", "./migrations", "Path to migrations directory")
}

func main() {
	flag.Parse()

	if dsn == "" {
		log.Fatal("missing --dsn")
	}

	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		log.Printf("failed opening DB: %v", err)
		return
	}
	defer db.Close()

	fmt.Println("Applying migrations...")
	if err := goose.Up(db, migrations); err != nil {
		log.Printf("migration failed: %v", err)
		return
	}

	fmt.Println("Migrations applied successfully")
}
