package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"gintugas/database"
	_ "gintugas/docs"
	routers "gintugas/modules"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // DRIVER PGX
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Gintugas API
// @version 1.0
// @description API untuk manajemen tugas dan proyek
// @host localhost:8080
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
	// Load env (jika lokal)
	godotenv.Load("config/.env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not found")
	}

	// Connect SQL (pgx)
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	fmt.Println("Berhasil koneksi ke database Koyeb")

	// Connect GORM
	gormDB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to create GORM connection:", err)
	}

	// Run Auto Migration SQL
	database.DBMigrate(db)

	InitiateRouter(db, gormDB)
}

func InitiateRouter(db *sql.DB, gormDB *gorm.DB) {
	router := gin.Default()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.GET("/api/health", healthCheck)
	routers.Initiator(router, db, gormDB)

	log.Printf("Server running on port %s", port)
	router.Run(":" + port)
}

// HealthCheck godoc
// @Summary Health check
// @Description Check API status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
