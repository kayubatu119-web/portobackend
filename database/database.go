package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed sql_migrations/*.sql
var dbMigrations embed.FS

var DbConnection *sql.DB

// database/migration.go - dengan logging
func DBMigrate(dbParam *sql.DB) {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: dbMigrations,
		Root:       "sql_migrations",
	}

	// Get pending migrations
	pending, err := migrate.GetMigrationRecords(dbParam, "postgres")
	if err != nil {
		log.Printf("Warning: Cannot get migration records: %v", err)
	} else {
		log.Printf("Existing migrations: %d", len(pending))
	}

	// Apply migrations
	n, errs := migrate.Exec(dbParam, "postgres", migrations, migrate.Up)
	if errs != nil {
		log.Printf("Migration failed: %v", errs)

		// Try individual migration
		migrate.SetTable("migrations")
		n, errs = migrate.ExecMax(dbParam, "postgres", migrations, migrate.Up, 1)
		if errs != nil {
			panic(fmt.Sprintf("Critical migration error: %v", errs))
		}
	}

	DbConnection = dbParam
	log.Printf("✅ Migration success, applied %d migrations!", n)
}
