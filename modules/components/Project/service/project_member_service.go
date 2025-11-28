package projectservice

import (
	repository "gintugas/modules/components/Project/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectMemberService struct {
	MemberRepo  *repository.ProjectMemberRepo
	ProjectRepo repository.Repository // Tambahkan ini untuk get project dengan tags
}

func NewProjectMemberService(memberRepo *repository.ProjectMemberRepo, projectRepo repository.Repository) *ProjectMemberService {
	return &ProjectMemberService{
		MemberRepo:  memberRepo,
		ProjectRepo: projectRepo,
	}
}

type AddTagRequest struct {
	TagID string `json:"tag_id" binding:"required,uuid"`
}

// AddTag menambahkan tag ke project
func (s *ProjectMemberService) AddTag(ctx *gin.Context) {
	projectID := ctx.Param("project_id")

	// Validasi UUID
	if _, err := uuid.Parse(projectID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return
	}

	var req AddTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi tag ID
	if _, err := uuid.Parse(req.TagID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID format"})
		return
	}

	// Cek apakah tag sudah ada
	exists, err := s.MemberRepo.IsProjectTag(projectID, req.TagID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tag already added to this project"})
		return
	}

	// Tambahkan tag
	if err := s.MemberRepo.AddTag(projectID, req.TagID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get project dengan tags untuk response
	pID, _ := uuid.Parse(projectID)
	project, err := s.ProjectRepo.GetProjekWithTagsRepository(pID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "Tag added successfully"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Tag added successfully",
		"data":    project,
	})
}

// RemoveTag menghapus tag dari project
func (s *ProjectMemberService) RemoveTag(ctx *gin.Context) {
	projectID := ctx.Param("project_id")
	tagID := ctx.Param("tag_id")

	// Validasi UUID
	if _, err := uuid.Parse(projectID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return
	}

	if _, err := uuid.Parse(tagID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID format"})
		return
	}

	if err := s.MemberRepo.RemoveTag(projectID, tagID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get project dengan tags untuk response
	pID, _ := uuid.Parse(projectID)
	project, err := s.ProjectRepo.GetProjekWithTagsRepository(pID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "Tag removed successfully"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Tag removed successfully",
		"data":    project,
	})
}

// GetProjectTags mendapatkan semua tags dalam project
func (s *ProjectMemberService) GetProjectTags(ctx *gin.Context) {
	projectID := ctx.Param("project_id")

	// Validasi UUID
	if _, err := uuid.Parse(projectID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return
	}

	tags, err := s.MemberRepo.GetProjectTags(projectID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": tags,
	})
}
