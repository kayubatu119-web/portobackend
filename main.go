package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"gintugas/database"
	_ "gintugas/docs"
	routers "gintugas/modules"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
)

// @title Gintugas API
// @version 1.0
// @description API untuk manajemen tugas dan proyek
// @host your-app.koyeb.app
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header menggunakan format: Bearer {token}

var (
	db     *sql.DB
	gormDB *gorm.DB
	err    error
)

func main() {
	// Load .env hanya untuk development
	// Di Koyeb, pakai environment variables
	if os.Getenv("GIN_MODE") != "release" {
		if err := godotenv.Load("config/.env"); err != nil {
			log.Println("Using environment variables (no .env file)")
		}
	}

	fmt.Println("\n=== ENV VARS ===")
	for _, env := range []string{
		"SUPABASE_URL",
		"SUPABASE_SERVICE_ROLE_KEY",
		"SUPABASE_STORAGE_BUCKET",
	} {
		val := os.Getenv(env)
		if val == "" {
			fmt.Printf("❌ %s: (empty)\n", env)
		} else {
			fmt.Printf("✅ %s: %s...\n", env, val[:min(10, len(val))])
		}
	}

	// Setup database
	db, gormDB = setupDatabase()
	defer db.Close()

	// Run migrations
	database.DBMigrate(db)

	// Start server
	InitiateRouter(db, gormDB)
}

func setupDatabase() (*sql.DB, *gorm.DB) {
	// Get database URL dengan SSL enabled
	dbURL := getDatabaseURL()

	log.Printf("Connecting to database with URL: %s", maskPassword(dbURL))

	// Open connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection dengan timeout
	err = pingWithTimeout(db, 10*time.Second)
	if err != nil {
		log.Printf("Database ping error: %v", err)
		log.Printf("Connection URL: %s", maskPassword(dbURL))
		log.Fatal("Failed to connect to database. Check SSL configuration.")
	}

	fmt.Println("✅ Berhasil Koneksi Ke Database")

	// Test SSL status
	testSSLStatus(db)

	// Setup GORM
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to create GORM connection:", err)
	}

	return db, gormDB
}

func getDatabaseURL() string {
	// Priority 1: DATABASE_URL dari environment (sudah include SSL)
	if url := os.Getenv("DATABASE_URL"); url != "" {
		// Pastikan ada sslmode=require
		if !strings.Contains(url, "sslmode=") {
			if strings.Contains(url, "?") {
				url += "&sslmode=require"
			} else {
				url += "?sslmode=require"
			}
		}
		return url
	}

	// Priority 2: Supabase dengan SSL
	// Password harus di-encode: @ -> %40, # -> %23
	encodedPass := "Bg3644aa%40%23"
	supabaseURL := fmt.Sprintf(
		"postgresql://postgres:%s@db.yiujndqqbacipqozosdm.supabase.co:5432/postgres?sslmode=require&sslmode=require&sslrootcert=system",
		encodedPass,
	)

	return supabaseURL
}

func testSSLStatus(db *sql.DB) {
	var sslStatus string
	err := db.QueryRow("SHOW ssl").Scan(&sslStatus)
	if err != nil {
		log.Printf("⚠️ Cannot check SSL status: %v", err)
	} else {
		log.Printf("🔒 SSL Status: %s", sslStatus)
	}

	// Check connection encryption
	var sslInfo string
	err = db.QueryRow("SELECT ssl_cipher()").Scan(&sslInfo)
	if err != nil {
		log.Printf("⚠️ Cannot check SSL cipher: %v", err)
	} else if sslInfo != "" {
		log.Printf("🔐 SSL Cipher: %s", sslInfo)
	} else {
		log.Println("❌ SSL NOT ENABLED on connection!")
	}
}

func maskPassword(url string) string {
	// Mask password untuk logging
	re := regexp.MustCompile(`password=[^&]*`)
	return re.ReplaceAllString(url, "password=****")
}

func pingWithTimeout(db *sql.DB, timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- db.Ping()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("database ping timeout after %v", timeout)
	}
}

func InitiateRouter(db *sql.DB, gormDB *gorm.DB) {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Get port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		// Check database connection
		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  "Database disconnected",
				"time":   time.Now().Format(time.RFC3339),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "gintugas-api",
			"time":    time.Now().Format(time.RFC3339),
			"version": "1.0",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Gintugas API",
			"version": "1.0",
			"docs":    "/swagger/index.html",
			"health":  "/health",
		})
	})

	// API routes
	routers.Initiator(router, db, gormDB)

	// Serve static files (uploads) - untuk Koyeb pakai external storage
	router.Static("/uploads", "./uploads")

	log.Printf("🚀 Server running on port %s", port)
	log.Println("📚 Swagger UI: http://localhost:" + port + "/swagger/index.html")

	// Bind ke 0.0.0.0 untuk Koyeb
	router.Run("0.0.0.0:" + port)
}
