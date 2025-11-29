package main

import (
	"database/sql"
	"fmt"
	_ "gintugas/docs"
	routers "gintugas/modules"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
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
	err = godotenv.Load("config/.env")
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	psqlInfo := fmt.Sprintf(`host=%s port=%s user=%s password=%s dbname=%s sslmode=disable`,
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGDATABASE"),
	)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Berhasil Koneksi Ke Database")

	gormDB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to create GORM connection:", err)
	}

	// database.DBMigrate(db)

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
	log.Println("Swagger UI available at: http://localhost:8080/swagger/index.html")
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
