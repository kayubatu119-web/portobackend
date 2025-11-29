package serviceroute

import (
	"gintugas/modules/components/experiences/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GormExpeHandler struct {
	expeService service.ExperiencesService
}

func NewGormExpeHandler(expeService service.ExperiencesService) *GormExpeHandler {
	return &GormExpeHandler{
		expeService: expeService,
	}
}

// ============================
// GORM HANDLERS
// ============================

func (c *GormExpeHandler) CreateExperiencesWithRelations(ctx *gin.Context) {
	experience, err := c.expeService.CreateExperienceWithRelations(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Experience with relations created successfully",
		"experience": experience,
	})
}

func (c *GormExpeHandler) GetExperiencesByIDWithRelations(ctx *gin.Context) {
	experience, err := c.expeService.GetExperienceByIDWithRelations(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Experience with relations retrieved successfully",
		"experience": experience,
	})
}

func (c *GormExpeHandler) UpdateExperiencesWithRelations(ctx *gin.Context) {
	experience, err := c.expeService.UpdateExperienceWithRelations(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Experience with relations updated successfully",
		"experience": experience,
	})
}

func (c *GormExpeHandler) DeleteExperiencesWithRelations(ctx *gin.Context) {
	if err := c.expeService.DeleteExperienceWithRelations(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Experience with relations deleted successfully",
	})
}

func (c *GormExpeHandler) GetAllExperiencesWithRelations(ctx *gin.Context) {
	experiences, err := c.expeService.GetAllExperiencesWithRelations(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "All experiences with relations retrieved successfully",
		"experiences": experiences,
	})
}
