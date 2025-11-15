package main

import (
	"flag"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		log.Fatalf("failed opening DB: %v", err)
	}
	defer db.Close()

	fmt.Println("Applying migrations...")
	if err := goose.Up(db, migrations); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Println("Migrations applied successfully")
}
