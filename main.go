package main

import (
	"context"
	"database/sql"
	"fmt"
	"gintugas/database"
	_ "gintugas/docs"
	routers "gintugas/modules"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool" // IMPORT INI
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, gormDB, pool, err := setupDatabaseWithPool()
	if err != nil {
		log.Fatal("Database setup failed:", err)
	}
	defer db.Close()
	defer pool.Close()

	database.DBMigrate(db)
	startServer(db, gormDB, port)
}

// METHOD 1: Gunakan pgxpool (RECOMMENDED)
func setupDatabaseWithPool() (*sql.DB, *gorm.DB, *pgxpool.Pool, error) {
	dbURL := getDatabaseURL()

	// Parse config untuk pool
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse config failed: %v", err)
	}

	// Configure pool
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Create pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create pool failed: %v", err)
	}

	// Test connection
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()

	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, nil, nil, fmt.Errorf("ping failed: %v", err)
	}

	// Convert pool ke database/sql
	sqlDB := stdlib.OpenDBFromPool(pool)

	// Setup GORM
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		sqlDB.Close()
		pool.Close()
		return nil, nil, nil, fmt.Errorf("gorm failed: %v", err)
	}

	log.Println("✅ Database pool connected")
	return sqlDB, gormDB, pool, nil
}

func getDatabaseURL() string {
	// Priority 1: DATABASE_URL
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	// Priority 2: Koyeb DB
	host := os.Getenv("PGHOST")
	user := os.Getenv("PGUSER")
	pass := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")
	port := os.Getenv("PGPORT")

	if host != "" && user != "" {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
			user, pass, host, port, dbname)
	}

	// Default Supabase (encoded password)
	return "postgresql://postgres:Bg3644aa%40%23@db.yiujndqqbacipqozosdm.supabase.co:5432/postgres?sslmode=require"
}

func startServer(db *sql.DB, gormDB *gorm.DB, port string) {
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Routes
	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "gintugas",
			"database":  "connected",
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Gintugas API",
			"version": "1.0",
			"health":  "/health",
			"docs":    "/swagger/index.html",
		})
	})

	// Your routes
	routers.Initiator(router, db, gormDB)

	// Swagger
	router.Static("/swagger", "./docs")

	addr := "0.0.0.0:" + port
	log.Printf("🚀 Server starting on %s", addr)
	router.Run(addr)
}
