package main

import (
	"flag"
	"log"
	"os"

	appmigrate "market-order-service/internal/app/migrate"
)

func main() {
	target := flag.String("target", "postgres", "migration target: postgres")
	dir := flag.String("dir", "", "migrations directory")
	flag.Parse()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN is required")
	}

	command := "up"
	if flag.NArg() > 0 {
		command = flag.Arg(0)
	}

	migrDir := *dir
	if migrDir == "" {
		switch *target {
		case "postgres":
			migrDir = "migrations/postgres"
		default:
			log.Fatalf("unknown target: %s", *target)
		}
	}

	if err := appmigrate.Run(dsn, migrDir, command); err != nil {
		log.Fatalf("migrate %s: %v", command, err)
	}
}
