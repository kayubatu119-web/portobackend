package modules

import (
	"database/sql"
	handlers "gintugas/modules/ServiceRoute"
	serviceroute "gintugas/modules/ServiceRoute"
	projectRPO "gintugas/modules/components/Project/repository"
	repositoryprojek "gintugas/modules/components/Project/repository"
	projectServsc "gintugas/modules/components/Project/service"
	"gintugas/modules/components/experiences/repo"
	"gintugas/modules/components/experiences/service"
	"log"
	"os"
	"path/filepath"
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
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	uploadBasePath := getUploadPath()

	// Create upload directories jika tidak ada (untuk development)
	if os.Getenv("GIN_MODE") != "release" {
		createUploadDirs(uploadBasePath)
	}

	// ============================
	// PROJECT DEPENDENCIES
	// ============================
	projectRepo := projectRPO.NewRepository(db)
	projectService := projectServsc.NewService(projectRepo, filepath.Join(uploadBasePath, "projects"))
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

	// Skills
	skillRepo := portfolioRepo.NewSkillRepository(gormDB)
	skillService := portfolioService.NewSkillService(skillRepo, filepath.Join(uploadBasePath, "skills"))
	skillHandler := handlers.NewSkillHandler(skillService)

	// Certificates
	certRepo := portfolioRepo.NewCertificateRepository(gormDB)
	certService := portfolioService.NewCertificateService(certRepo, filepath.Join(uploadBasePath, "certificates"))
	certHandler := handlers.NewCertificateHandler(certService)

	// Education
	eduRepo := portfolioRepo.NewEducationRepository(gormDB)
	eduService := portfolioService.NewEducationService(eduRepo)
	eduHandler := handlers.NewEducationHandler(eduService)

	// Testimonials
	testRepo := portfolioRepo.NewTestimonialRepository(gormDB)
	testService := portfolioService.NewTestimonialService(testRepo)
	testHandler := handlers.NewTestimonialHandler(testService)

	// Blog
	blogRepo := portfolioRepo.NewBlogRepository(gormDB)
	blogService := portfolioService.NewBlogService(blogRepo)
	blogHandler := handlers.NewBlogHandler(blogService)

	// Sections
	sectionRepo := portfolioRepo.NewSectionRepository(gormDB)
	sectionService := portfolioService.NewSectionService(sectionRepo)
	sectionHandler := handlers.NewSectionHandler(sectionService)

	// Social Links
	socialLinkRepo := portfolioRepo.NewSocialLinkRepository(gormDB)
	socialLinkService := portfolioService.NewSocialLinkService(socialLinkRepo)
	socialLinkHandler := handlers.NewSocialLinkHandler(socialLinkService)

	// Settings
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
	// SERVE STATIC FILES
	// ============================
	if os.Getenv("GIN_MODE") != "release" {
		router.Static("/uploads", uploadBasePath)
		log.Printf("📁 Serving static files from: %s", uploadBasePath)
	} else {
		log.Println("ℹ️  In production mode, using external storage for uploads")
	}
}

func getUploadPath() string {
	// Di Koyeb, pakai /tmp karena ephemeral
	// Atau pakai external storage (S3, Cloudinary, dll)
	if os.Getenv("GIN_MODE") == "release" {
		// Untuk Koyeb, pakai /tmp atau volume
		if path := os.Getenv("UPLOAD_PATH"); path != "" {
			return path
		}
		return "/tmp/uploads"
	}

	// Untuk development, pakai local folder
	return "./uploads"
}

func createUploadDirs(basePath string) {
	dirs := []string{
		"projects",
		"skills",
		"certificates",
		// tambah folder lainnya
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
