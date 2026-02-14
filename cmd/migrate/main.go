package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"url-shortener-go/config"
)

func main() {
	upFlag := flag.Bool("up", false, "Apply all up migrations")
	downFlag := flag.Bool("down", false, "Apply all down migrations")
	versionFlag := flag.Int("version", 0, "Migrate to a specific version")
	flag.Parse()

	if (*upFlag && *downFlag) || (*upFlag && *versionFlag != 0) || (*downFlag && *versionFlag != 0) {
		log.Fatal("Flags --up, --down, and --version are mutually exclusive.")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := postgres.WithInstance(
		mustOpenDB(cfg.GetPostgresConnString()),
		&postgres.Config{},
	)
	if err != nil {
		log.Fatalf("Error creating PostgreSQL driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		cfg.MigrationsPath,
		"postgres",
		db,
	)
	if err != nil {
		log.Fatalf("Error creating migrate instance: %v", err)
	}

	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("Source close error: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("Database close error: %v", dbErr)
		}
	}()

	if *upFlag {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error applying migrations: %v", err)
		}
		fmt.Println("Migrations applied successfully!")
		return
	}

	if *downFlag {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error reverting migrations: %v", err)
		}
		fmt.Println("Migrations reverted successfully!")
		return
	}

	if *versionFlag != 0 {
		if err := m.Migrate(uint(*versionFlag)); err != nil {
			log.Fatalf("Error migrating to version %d: %v", *versionFlag, err)
		}
		fmt.Printf("Migrated to version %d successfully!\n", *versionFlag)
		return
	}

	flag.Usage()
	os.Exit(1)
}

func mustOpenDB(connStr string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	return db
}
