package modules

import (
	"database/sql"
	"fmt"
	handlers "gintugas/modules/ServiceRoute"
	serviceroute "gintugas/modules/ServiceRoute"
	projectRPO "gintugas/modules/components/Project/repository"
	repositoryprojek "gintugas/modules/components/Project/repository"
	projectServsc "gintugas/modules/components/Project/service"
	"gintugas/modules/components/experiences/repo"
	"gintugas/modules/components/experiences/service"
	"gintugas/modules/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Import portfolio components
	portfolioRepo "gintugas/modules/components/all/repo"
	portfolioService "gintugas/modules/components/all/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func Initiator(router *gin.Engine, db *sql.DB, gormDB *gorm.DB) {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	uploadBasePath := getUploadPath()

	// ============================
	// CREATE UPLOAD SERVICES
	// ============================
	fmt.Println("=== STORAGE CONFIGURATION ===")

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucket := os.Getenv("SUPABASE_STORAGE_BUCKET")
	uploadProvider := os.Getenv("UPLOAD_PROVIDER")

	fmt.Printf("SUPABASE_URL: %s\n", supabaseURL)

	// ⭐ PERBAIKAN: Cek panjang key sebelum menggunakan strings.Repeat
	if supabaseServiceKey != "" && len(supabaseServiceKey) > 10 {
		fmt.Printf("SUPABASE_SERVICE_ROLE_KEY: %s\n",
			strings.Repeat("*", len(supabaseServiceKey)-10)+supabaseServiceKey[len(supabaseServiceKey)-10:])
	} else if supabaseServiceKey != "" {
		fmt.Printf("SUPABASE_SERVICE_ROLE_KEY: %s\n", strings.Repeat("*", len(supabaseServiceKey)))
	} else {
		fmt.Printf("SUPABASE_SERVICE_ROLE_KEY: (not set)\n")
	}

	fmt.Printf("SUPABASE_STORAGE_BUCKET: %s\n", bucket)
	fmt.Printf("UPLOAD_PROVIDER: %s\n", uploadProvider)

	// ⭐ PERBAIKAN: Debug semua env variables
	fmt.Println("\n=== ENVIRONMENT VARIABLES ===")
	envVars := []string{
		"SUPABASE_URL",
		"SUPABASE_ANON_KEY",
		"SUPABASE_SERVICE_ROLE_KEY",
		"SUPABASE_STORAGE_BUCKET",
		"UPLOAD_PROVIDER",
		"GIN_MODE",
	}

	for _, env := range envVars {
		value := os.Getenv(env)
		if value == "" {
			fmt.Printf("⚠️  %s: (not set)\n", env)
		} else {
			fmt.Printf("✅ %s: %s\n", env, value[:min(len(value), 50)])
		}
	}

	var supabaseUploadService *utils.SupabaseUploadService

	// Setup Supabase Storage
	if supabaseURL != "" && supabaseServiceKey != "" {
		fmt.Println("\n🔄 Initializing Supabase Storage...")
		supabaseUploadService = utils.NewSupabaseUploadService(supabaseURL, supabaseServiceKey, bucket)
		uploadProvider = "supabase"
		fmt.Println("✅ Supabase Storage initialized")
	} else {
		uploadProvider = "local"
		fmt.Println("\n⚠️  Supabase Storage not configured, using local storage")
		if supabaseURL == "" {
			fmt.Println("   Reason: SUPABASE_URL is empty")
		}
		if supabaseServiceKey == "" {
			fmt.Println("   Reason: SUPABASE_SERVICE_ROLE_KEY is empty")
		}
	}

	// ============================
	// PROJECT DEPENDENCIES
	// ============================
	projectRepo := projectRPO.NewRepository(db)

	var projectService projectServsc.Service
	if uploadProvider == "supabase" && supabaseUploadService != nil {
		// Create Supabase wrapper
		supabaseWrapper := projectServsc.NewSupabaseUploadWrapper(supabaseUploadService)
		// Use existing NewService but pass upload wrapper
		projectService = projectServsc.NewServiceWithUpload(projectRepo, supabaseWrapper, "projects")
		fmt.Println("📁 Project Service: Using Supabase Storage")
	} else {
		// Use local storage
		localPath := filepath.Join(uploadBasePath, "projects")
		projectService = projectServsc.NewService(projectRepo, localPath)
		fmt.Println("📁 Project Service: Using Local Storage")
	}
	projectHandler := handlers.NewProjectHandler(projectService)

	memberRepo := repositoryprojek.NewProjectMemberRepo(gormDB)
	memberService := projectServsc.NewProjectMemberService(memberRepo, projectRepo)

	tagsrepo := projectRPO.NewTagsRepository(gormDB)
	tagsService := projectServsc.NewTaskService(tagsrepo)
	tagsHandler := handlers.NewTagsHandler(tagsService)

	// ============================
	// EXPERIENCE DEPENDENCIES
	// ============================
	expeRepo := repo.NewExpeGormRepository(gormDB)
	expeService := service.NewExpeService(expeRepo)
	expeHandler := serviceroute.NewGormExpeHandler(expeService)

	// ============================
	// PORTFOLIO DEPENDENCIES
	// ============================

	// Skills with upload service
	skillRepo := portfolioRepo.NewSkillRepository(gormDB)
	var skillService portfolioService.SkillService
	if uploadProvider == "supabase" && supabaseUploadService != nil {
		supabaseWrapper := portfolioService.NewSupabaseUploadWrapper(supabaseUploadService)
		skillService = portfolioService.NewSkillServiceWithUpload(skillRepo, supabaseWrapper, "skills")
	} else {
		localPath := filepath.Join(uploadBasePath, "skills")
		skillService = portfolioService.NewSkillService(skillRepo, localPath)
	}
	skillHandler := handlers.NewSkillHandler(skillService)

	// Certificates with upload service
	certRepo := portfolioRepo.NewCertificateRepository(gormDB)
	var certService portfolioService.CertificateService
	if uploadProvider == "supabase" && supabaseUploadService != nil {
		supabaseWrapper := portfolioService.NewSupabaseUploadWrapper(supabaseUploadService)
		certService = portfolioService.NewCertificateServiceWithUpload(certRepo, supabaseWrapper, "certificates")
	} else {
		localPath := filepath.Join(uploadBasePath, "certificates")
		certService = portfolioService.NewCertificateService(certRepo, localPath)
	}
	certHandler := handlers.NewCertificateHandler(certService)

	// Education (no upload needed)
	eduRepo := portfolioRepo.NewEducationRepository(gormDB)
	eduService := portfolioService.NewEducationService(eduRepo)
	eduHandler := handlers.NewEducationHandler(eduService)

	// Testimonials (no upload needed)
	testRepo := portfolioRepo.NewTestimonialRepository(gormDB)
	testService := portfolioService.NewTestimonialService(testRepo)
	testHandler := handlers.NewTestimonialHandler(testService)

	// Blog (no upload needed)
	blogRepo := portfolioRepo.NewBlogRepository(gormDB)
	blogService := portfolioService.NewBlogService(blogRepo)
	blogHandler := handlers.NewBlogHandler(blogService)

	// Sections (no upload needed)
	sectionRepo := portfolioRepo.NewSectionRepository(gormDB)
	sectionService := portfolioService.NewSectionService(sectionRepo)
	sectionHandler := handlers.NewSectionHandler(sectionService)

	// Social Links (no upload needed)
	socialLinkRepo := portfolioRepo.NewSocialLinkRepository(gormDB)
	socialLinkService := portfolioService.NewSocialLinkService(socialLinkRepo)
	socialLinkHandler := handlers.NewSocialLinkHandler(socialLinkService)

	// Settings (no upload needed)
	settingRepo := portfolioRepo.NewSettingRepository(gormDB)
	settingService := portfolioService.NewSettingService(settingRepo)
	settingHandler := handlers.NewSettingHandler(settingService)

	// ============================
	// SWAGGER
	// ============================
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ============================
	// API ROUTES
	// ============================
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "ok",
				"service":   "gintugas-api",
				"timestamp": time.Now().Unix(),
				"upload":    uploadProvider,
			})
		})

		// ============================
		// PROJECT ROUTES
		// ============================
		projectRoutes := api.Group("/v1/projects")
		{
			projectRoutes.GET("", projectHandler.GetAllProjects)
			projectRoutes.GET("/:id", projectHandler.GetProject)
			projectRoutes.POST("/with-image", projectHandler.CreateProjectWithImage)
			projectRoutes.PUT("/:id", projectHandler.UpdateProject)
			projectRoutes.DELETE("/:id", projectHandler.DeleteProject)
		}

		projects := api.Group("/projects")
		{
			projects.POST("/:project_id/tags", memberService.AddTag)
			projects.DELETE("/:project_id/tags/:tag_id", memberService.RemoveTag)
			projects.GET("/:project_id/tags", memberService.GetProjectTags)
		}

		tags := api.Group("/v1/tags")
		{
			tags.POST("", tagsHandler.CreateTags)
			tags.GET("", projectHandler.GetAllTags)
		}

		// ============================
		// EXPERIENCE ROUTES
		// ============================
		expeRoutes := api.Group("/v1")
		{
			expeRoutes.POST("/experiences/with-relations", expeHandler.CreateExperiencesWithRelations)
			expeRoutes.GET("/experiences/with-relations", expeHandler.GetAllExperiencesWithRelations)
			expeRoutes.GET("/experiences/with-relations/:id", expeHandler.GetExperiencesByIDWithRelations)
			expeRoutes.PUT("/experiences/with-relations/:id", expeHandler.UpdateExperiencesWithRelations)
			expeRoutes.DELETE("/experiences/with-relations/:id", expeHandler.DeleteExperiencesWithRelations)
		}

		// ============================
		// PORTFOLIO ROUTES
		// ============================
		v1 := api.Group("/v1")

		// SKILLS ROUTES
		skills := v1.Group("/skills")
		{
			skills.POST("", skillHandler.Create)
			skills.POST("/with-icon", skillHandler.CreateWithIcon)
			skills.PUT("/:id/with-icon", skillHandler.UpdateWithIcon)
			skills.GET("", skillHandler.GetAll)
			skills.GET("/featured", skillHandler.GetFeatured)
			skills.GET("/category/:category", skillHandler.GetByCategory)
			skills.GET("/:id", skillHandler.GetByID)
			skills.PUT("/:id", skillHandler.Update)
			skills.DELETE("/:id", skillHandler.Delete)
		}

		// CERTIFICATES ROUTES
		certificates := v1.Group("/certificates")
		{
			certificates.POST("", certHandler.Create)
			certificates.POST("/with-image", certHandler.CreateWithImage)
			certificates.GET("", certHandler.GetAll)
			certificates.GET("/:id", certHandler.GetByID)
			certificates.PUT("/:id", certHandler.Update)
			certificates.DELETE("/:id", certHandler.Delete)
		}

		// EDUCATION ROUTES
		education := v1.Group("/education")
		{
			education.POST("", eduHandler.CreateWithAchievements)
			education.GET("", eduHandler.GetAllWithAchievements)
			education.GET("/:id", eduHandler.GetByIDWithAchievements)
			education.PUT("/:id", eduHandler.UpdateWithAchievements)
			education.DELETE("/:id", eduHandler.DeleteWithAchievements)
		}

		// TESTIMONIALS ROUTES
		testimonials := v1.Group("/testimonials")
		{
			testimonials.POST("", testHandler.Create)
			testimonials.GET("", testHandler.GetAll)
			testimonials.GET("/featured", testHandler.GetFeatured)
			testimonials.GET("/status/:status", testHandler.GetByStatus)
			testimonials.GET("/:id", testHandler.GetByID)
			testimonials.PUT("/:id", testHandler.Update)
			testimonials.DELETE("/:id", testHandler.Delete)
		}

		// BLOG ROUTES
		blog := v1.Group("/blog")
		{
			blog.POST("", blogHandler.CreateWithTags)
			blog.GET("", blogHandler.GetAllWithTags)
			blog.GET("/published", blogHandler.GetPublishedWithTags)
			blog.GET("/tags", blogHandler.GetAllTags)
			blog.GET("/:id", blogHandler.GetByIDWithTags)
			blog.GET("/slug/:slug", blogHandler.GetBySlugWithTags)
			blog.PUT("/:id", blogHandler.UpdateWithTags)
			blog.DELETE("/:id", blogHandler.DeleteWithTags)
		}

		// SECTIONS ROUTES
		sections := v1.Group("/sections")
		{
			sections.POST("", sectionHandler.Create)
			sections.GET("", sectionHandler.GetAll)
			sections.DELETE("/:id", sectionHandler.Delete)
		}

		// SOCIAL LINKS ROUTES
		socialLinks := v1.Group("/social-links")
		{
			socialLinks.POST("", socialLinkHandler.Create)
			socialLinks.GET("", socialLinkHandler.GetAll)
			socialLinks.DELETE("/:id", socialLinkHandler.Delete)
		}

		// SETTINGS ROUTES
		settings := v1.Group("/settings")
		{
			settings.POST("", settingHandler.Create)
			settings.GET("", settingHandler.GetAll)
			settings.DELETE("/:id", settingHandler.Delete)
		}
	}

	// ============================
	// SERVE STATIC FILES (Development only)
	// ============================
	if os.Getenv("GIN_MODE") != "release" && uploadProvider == "local" {
		router.Static("/uploads", uploadBasePath)
		log.Printf("📁 Serving static files from: %s", uploadBasePath)
	} else {
		log.Printf("ℹ️  Using %s storage for uploads", uploadProvider)
	}
}

func getUploadPath() string {
	if os.Getenv("GIN_MODE") == "release" {
		if path := os.Getenv("UPLOAD_PATH"); path != "" {
			return path
		}
		return "/tmp/uploads"
	}
	return "./uploads"
}

func createUploadDirs(basePath string) {
	dirs := []string{
		"projects",
		"skills",
		"certificates",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(basePath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			log.Printf("Warning: Cannot create upload directory %s: %v", fullPath, err)
		} else {
			log.Printf("Created upload directory: %s", fullPath)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
