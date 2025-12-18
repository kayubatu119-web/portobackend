package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
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
	if os.Getenv("GIN_MODE") != "release" {
		if err := godotenv.Load("config/.env"); err != nil {
			fmt.Println("Using environment variables (no .env file)")
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🚀 GINTUGAS API STARTING")
	fmt.Println(strings.Repeat("=", 50))

	// Show important env vars
	fmt.Println("\n📋 ENVIRONMENT VARIABLES:")
	envVars := []string{
		"SUPABASE_URL",
		"SUPABASE_STORAGE_BUCKET",
		"UPLOAD_PROVIDER",
		"GIN_MODE",
		"PORT",
	}

	for _, env := range envVars {
		val := os.Getenv(env)
		if val == "" {
			fmt.Printf("   ❌ %s: (not set)\n", env)
		} else {
			fmt.Printf("   ✅ %s: %s\n", env, val)
		}
	}

	// Setup database
	db, gormDB, dbConnected := setupDatabase()

	if dbConnected && db != nil {
		defer func() {
			fmt.Println("🔌 Closing database connection...")
			db.Close()
		}()

		fmt.Println("\n🗄️ Running database migrations...")
		err := database.DBMigrate(db)
		if err != nil {
			fmt.Printf("⚠️ Database migrations failed: %v\n", err)
			fmt.Println("⚠️ Continuing without migrations...")
		} else {
			fmt.Println("✅ Database migrations completed")
		}
	} else {
		fmt.Println("\n⚠️ Skipping database operations (database not connected)")
		fmt.Println("⚠️ File uploads to Supabase will still work")
	}

	// Start server
	InitiateRouter(db, gormDB)
}

func setupDatabase() (*sql.DB, *gorm.DB, bool) {
	// Get database URL dengan force IPv4
	dbURL := getDatabaseURL()

	fmt.Println("\n🔌 Setting up database connection...")
	fmt.Printf("   Connection URL: %s\n", maskPassword(dbURL))

	// ⭐ PERBAIKAN: Coba koneksi dengan timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", dbURL)
	if os.Getenv("SKIP_DATABASE") == "true" {
		fmt.Println("⚠️ SKIP_DATABASE=true - Skipping database connection")
		return nil, nil, false
	}

	// Test connection dengan context timeout
	fmt.Println("   Testing database connection...")
	err = db.PingContext(ctx)
	if err != nil {
		fmt.Printf("⚠️ Database connection failed: %v\n", err)
		fmt.Println("⚠️ This is OK - running in UPLOAD-ONLY mode")
		fmt.Println("⚠️ Database operations will fail, but file uploads will work")

		// ⭐ PERBAIKAN: Close koneksi yang gagal dan return nil
		db.Close()
		return nil, nil, false
	}

	fmt.Println("✅ Database connected successfully")

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Setup GORM
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		fmt.Printf("⚠️ GORM setup failed: %v\n", err)
		// Database connected, tapi GORM gagal
		return db, nil, true
	}

	fmt.Println("✅ GORM initialized")

	// Test Supabase storage connection
	testSupabaseUpload()

	return db, gormDB, true
}

func testSupabaseUpload() {
	fmt.Println("\n🔧 Testing Supabase Storage...")

	supabaseURL := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucket := os.Getenv("SUPABASE_STORAGE_BUCKET")

	if supabaseURL == "" || apiKey == "" {
		fmt.Println("⚠️ Supabase configuration missing")
		return
	}

	fmt.Printf("   URL: %s\n", supabaseURL)
	fmt.Printf("   Bucket: %s\n", bucket)
	fmt.Printf("   API Key available: %v\n", apiKey != "")

	// Test 1: Check bucket access
	testBucketAccess(supabaseURL, apiKey, bucket)

	// Test 2: Test upload with unique filename
	testUploadWithUniqueName(supabaseURL, apiKey, bucket)
}

func testBucketAccess(supabaseURL, apiKey, bucket string) {
	fmt.Println("\n📦 Testing bucket access...")

	url := fmt.Sprintf("%s/storage/v1/bucket/%s", supabaseURL, bucket)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("apikey", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Bucket access test failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		fmt.Println("✅ Bucket is accessible")
	} else {
		fmt.Printf("⚠️ Bucket access returned status %d: %s\n", resp.StatusCode, string(body))
	}
}

func testUploadWithUniqueName(supabaseURL, apiKey, bucket string) {
	fmt.Println("\n📤 Testing upload with unique filename...")

	// Buat file test di memory
	testContent := []byte("Test file content for Supabase upload")
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("test-%d.txt", timestamp)
	storagePath := fmt.Sprintf("debug/%s", filename)

	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s",
		strings.TrimSuffix(supabaseURL, "/"),
		bucket,
		storagePath,
	)

	fmt.Printf("   Upload URL: %s\n", uploadURL)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(testContent))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Cache-Control", "public, max-age=31536000")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Upload failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("   Status: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
			supabaseURL, bucket, storagePath)
		fmt.Printf("✅ Upload test successful!\n")
		fmt.Printf("   Public URL: %s\n", publicURL)

		// Test access
		testPublicAccess(publicURL)
	} else {
		fmt.Printf("❌ Upload test failed: %s\n", string(body))
	}
}

func testPublicAccess(url string) {
	fmt.Println("\n🔗 Testing public access...")

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Public access failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ File is publicly accessible!")
	} else {
		fmt.Printf("⚠️ Public access returned status %d\n", resp.StatusCode)
	}
}

func getDatabaseURL() string {
	// Priority 1: DATABASE_URL dari environment
	if url := os.Getenv("DATABASE_URL"); url != "" {
		url = strings.ReplaceAll(url, "@#", "%40%23")
		if !strings.Contains(url, "sslmode=") {
			if strings.Contains(url, "?") {
				url += "&sslmode=require"
			} else {
				url += "?sslmode=require"
			}
		}
		return url
	}

	// Priority 2: Coba dengan parameter terpisah untuk IPv4
	// Supabase biasanya butuh SSL
	password := "Bg3644aa%40%23"

	// ⭐ PERBAIKAN: Format connection string yang lebih baik
	return fmt.Sprintf(
		"postgresql://postgres:%s@db.yiujndqqbacipqozosdm.supabase.co:5432/postgres?sslmode=require&connect_timeout=10",
		password,
	)
}

func maskPassword(url string) string {
	re := regexp.MustCompile(`password=[^& ]*`)
	return re.ReplaceAllString(url, "password=****")
}

func InitiateRouter(db *sql.DB, gormDB *gorm.DB) {
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Recovery middleware
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("🚨 PANIC RECOVERED: %v", recovered)
		c.JSON(500, gin.H{
			"error":   "Internal server error",
			"message": "Something went wrong",
		})
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Health check yang lebih informatif
	router.GET("/health", func(c *gin.Context) {
		dbStatus := "not_configured"
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			if err := db.PingContext(ctx); err == nil {
				dbStatus = "connected"
			} else {
				dbStatus = "disconnected"
			}
		}

		c.JSON(200, gin.H{
			"status":          "ok",
			"service":         "gintugas-api",
			"timestamp":       time.Now().Unix(),
			"version":         "1.0",
			"database":        dbStatus,
			"upload_provider": os.Getenv("UPLOAD_PROVIDER"),
			"environment":     os.Getenv("GIN_MODE"),
			"capabilities": map[string]bool{
				"file_upload": os.Getenv("UPLOAD_PROVIDER") == "supabase",
				"database":    dbStatus == "connected",
			},
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Gintugas API",
			"version": "1.0",
			"status":  "running",
			"endpoints": map[string]string{
				"health":      "/health",
				"docs":        "/swagger/index.html",
				"upload_test": "/api/test-upload",
				"projects":    "/api/v1/projects",
			},
		})
	})

	// API routes
	routers.Initiator(router, db, gormDB)

	log.Printf("🚀 Server running on port %s", port)
	router.Run("0.0.0.0:" + port)
}
