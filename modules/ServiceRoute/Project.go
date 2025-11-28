package serviceroute

import (
	projectservice "gintugas/modules/components/Project/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService projectservice.Service
}

func NewProjectHandler(projectService projectservice.Service) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

type TagsHandler struct {
	tagsService projectservice.TagsService
}

func NewTagsHandler(tagsService projectservice.TagsService) *TagsHandler {
	return &TagsHandler{
		tagsService: tagsService,
	}
}

func (h *ProjectHandler) CreateProjectWithImage(c *gin.Context) {
	project, err := h.projectService.CreateProjekWithImageService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project created successfully",
		"data":    project,
	})
}

func (h *ProjectHandler) GetAllProjects(c *gin.Context) {
	projects, err := h.projectService.GetAllProjekService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": projects,
	})
}

func (h *ProjectHandler) GetAllTags(c *gin.Context) {
	projects, err := h.projectService.GetAllTagsService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": projects,
	})
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	project, err := h.projectService.GetProjekService(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": project,
	})
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	project, err := h.projectService.UpdateProjekService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project updated successfully",
		"data":    project,
	})
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	err := h.projectService.DeleteProjekService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}

func (c *TagsHandler) CreateTags(ctx *gin.Context) {
	Tags, err := c.tagsService.CreateTags(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Tags created successfully",
		"Tags":    Tags,
	})
}
