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

func DBMigrate(dbParam *sql.DB) error {
	// ⭐ PERBAIKAN: Cek database connection dulu
	if dbParam == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Test connection
	if err := dbParam.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: dbMigrations,
		Root:       "sql_migrations",
	}

	n, errs := migrate.Exec(dbParam, "postgres", migrations, migrate.Up)
	if errs != nil {
		// ⭐ PERBAIKAN: Return error, jangan panic
		return fmt.Errorf("migration failed: %v", errs)
	}

	DbConnection = dbParam

	log.Printf("✅ Migration success, applied %d migrations!\n", n)
	return nil
}
