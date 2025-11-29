package projectservice

import (
	"errors"
	"fmt"
	. "gintugas/modules/components/Project/model"
	. "gintugas/modules/components/Project/repository"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	GetAllTagsService(ctx *gin.Context) (result []ProjectTag, err error)
	GetAllProjekService(ctx *gin.Context) ([]Project, error)
	GetProjekService(ctx *gin.Context) (Project, error)
	UpdateProjekService(ctx *gin.Context) (Project, error)
	DeleteProjekService(ctx *gin.Context) error
	CreateProjekWithImageService(ctx *gin.Context) (Project, error)
}

type TagsService interface {
	CreateTags(ctx *gin.Context) (*TagResponse, error)
}

type projectService struct {
	repository Repository
	uploadPath string
}

func NewService(repository Repository, uploadPath string) Service {
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("Warning: gagal membuat folder upload: %v\n", err)
	}
	return &projectService{
		repository: repository,
		uploadPath: uploadPath,
	}
}

type tagsService struct {
	tagsRepo TagsRepository
}

func NewTaskService(tagsRepo TagsRepository) TagsService {
	return &tagsService{
		tagsRepo: tagsRepo,
	}
}

func (s *projectService) validateFile(file *multipart.FileHeader) error {
	// Ukuran file 10MB
	maxSize := int64(10 * 1024 * 1024)
	if file.Size > maxSize {
		return errors.New("ukuran file maksimal 10MB")
	}

	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		return errors.New("tipe file tidak diizinkan. File yang diizinkan: jpg, jpeg, png, webp")
	}

	return nil
}

func (s *projectService) CreateProjekWithImageService(ctx *gin.Context) (Project, error) {
	var form ProjectForm

	// Bind form data
	if err := ctx.ShouldBind(&form); err != nil {
		return Project{}, fmt.Errorf("gagal binding data: %v", err)
	}

	// Validasi required fields
	if form.Title == "" {
		return Project{}, errors.New("judul projek harus diisi")
	}

	if form.Description == "" {
		return Project{}, errors.New("deskripsi projek harus diisi")
	}

	if form.CodeURL == "" {
		return Project{}, errors.New("URL kode projek harus diisi")
	}

	// Handle file upload
	file, err := ctx.FormFile("image")

	// Tambahkan logging untuk debugging
	if err != nil {
		if err == http.ErrMissingFile {
			fmt.Println("⚠️ Warning: No file uploaded (missing file)")
		} else {
			fmt.Printf("❌ Error getting file: %v\n", err)
			return Project{}, fmt.Errorf("gagal mengambil file: %v", err)
		}
	}

	imageURL := "#"
	if file != nil {
		fmt.Printf("✅ File received: %s, Size: %d bytes, Content-Type: %s\n",
			file.Filename, file.Size, file.Header.Get("Content-Type"))

		// Validasi file
		if err := s.validateFile(file); err != nil {
			fmt.Printf("❌ File validation failed: %v\n", err)
			return Project{}, err
		}

		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("project_%s%s", uuid.New().String(), ext)
		filePath := filepath.Join(s.uploadPath, fileName)

		fmt.Printf("📁 Saving file to: %s\n", filePath)

		// Pastikan folder upload exists
		if err := os.MkdirAll(s.uploadPath, 0755); err != nil {
			return Project{}, fmt.Errorf("gagal membuat folder upload: %v", err)
		}

		// Simpan file
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			fmt.Printf("❌ Failed to save file: %v\n", err)
			return Project{}, fmt.Errorf("gagal menyimpan file: %v", err)
		}

		// Verifikasi file tersimpan
		if fileInfo, err := os.Stat(filePath); os.IsNotExist(err) {
			return Project{}, fmt.Errorf("file gagal disimpan: %v", err)
		} else {
			fmt.Printf("✅ File saved successfully, size: %d bytes\n", fileInfo.Size())
		}

		// Set image URL
		imageURL = "/uploads/projects/" + fileName
		fmt.Printf("🔗 Image URL set to: %s\n", imageURL)
	} else {
		fmt.Println("ℹ️ No file uploaded, using default image URL")
	}

	// Set default values
	if form.Status == "" {
		form.Status = "published"
	}
	if form.DemoURL == "" {
		form.DemoURL = "#"
	}

	// Convert form to Project entity
	project := Project{
		Title:        form.Title,
		Description:  form.Description,
		ImageURL:     imageURL,
		DemoURL:      form.DemoURL,
		CodeURL:      form.CodeURL,
		DisplayOrder: form.DisplayOrder,
		IsFeatured:   form.IsFeatured,
		Status:       form.Status,
	}

	result, err := s.repository.CreateProjekRepository(project)
	if err != nil {
		// Cleanup file jika gagal menyimpan data
		if file != nil && imageURL != "#" {
			fileToDelete := filepath.Join(s.uploadPath, filepath.Base(imageURL))
			os.Remove(fileToDelete)
			fmt.Printf("🗑️ Cleaned up file: %s\n", fileToDelete)
		}
		return Project{}, fmt.Errorf("gagal menyimpan data projek: %v", err)
	}

	fmt.Printf("✅ Project created successfully with ID: %s\n", result.ID)
	return result, nil
}

func (s *projectService) GetAllTagsService(ctx *gin.Context) (result []ProjectTag, err error) {
	Tags, err := s.repository.GetAllTagsRepository()
	if err != nil {
		return nil, errors.New("gagal mengambil data Tags: " + err.Error())
	}

	return Tags, nil
}

func (s *projectService) GetAllProjekService(ctx *gin.Context) ([]Project, error) {
	// Check query parameter for with_tags
	withTags := ctx.Query("with_tags")

	if withTags == "true" {
		return s.repository.GetAllProjekWithTagsRepository()
	}

	return s.repository.GetAllProjekRepository()
}

func (s *projectService) GetProjekService(ctx *gin.Context) (Project, error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return Project{}, errors.New("ID projek tidak valid")
	}

	// Check query parameter for with_tags
	withTags := ctx.Query("with_tags")

	if withTags == "true" {
		return s.repository.GetProjekWithTagsRepository(id)
	}

	return s.repository.GetProjekRepository(id)
}

// Service dengan struct binding
func (s *projectService) UpdateProjekService(ctx *gin.Context) (Project, error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return Project{}, errors.New("ID projek tidak valid")
	}

	// Check if project exists
	existingProject, err := s.repository.GetProjekRepository(id)
	if err != nil {
		return Project{}, errors.New("projek tidak ditemukan")
	}

	// Simpan image URL lama
	oldImageURL := existingProject.ImageURL

	// Handle file upload
	file, err := ctx.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		return Project{}, fmt.Errorf("gagal mengambil file: %v", err)
	}

	imageURL := existingProject.ImageURL
	if file != nil {
		fmt.Printf("✅ File received for update: %s\n", file.Filename)

		// Validasi file
		if err := s.validateFile(file); err != nil {
			return Project{}, err
		}

		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("project_%s%s", uuid.New().String(), ext)
		filePath := filepath.Join(s.uploadPath, fileName)

		// Simpan file baru
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			return Project{}, fmt.Errorf("gagal menyimpan file: %v", err)
		}

		imageURL = "/uploads/projects/" + fileName

		// Hapus file lama jika bukan default
		if oldImageURL != "" && oldImageURL != "#" {
			oldFileName := filepath.Base(oldImageURL)
			oldFilePath := filepath.Join(s.uploadPath, oldFileName)
			os.Remove(oldFilePath) // Ignore error
			fmt.Printf("🗑️ Deleted old file: %s\n", oldFilePath)
		}
	}

	// Bind form data
	var form ProjectUpdateForm
	if err := ctx.ShouldBind(&form); err != nil {
		// Cleanup file baru jika binding gagal
		if file != nil {
			newFileName := filepath.Base(imageURL)
			newFilePath := filepath.Join(s.uploadPath, newFileName)
			os.Remove(newFilePath)
		}
		return Project{}, fmt.Errorf("gagal binding data: %v", err)
	}

	// Update fields yang ada nilainya
	if form.Title != "" {
		existingProject.Title = form.Title
	}
	if form.Description != "" {
		existingProject.Description = form.Description
	}
	if form.DemoURL != "" {
		existingProject.DemoURL = form.DemoURL
	}
	if form.CodeURL != "" {
		existingProject.CodeURL = form.CodeURL
	}
	if form.DisplayOrder != 0 {
		existingProject.DisplayOrder = form.DisplayOrder
	}
	existingProject.IsFeatured = form.IsFeatured
	if form.Status != "" {
		existingProject.Status = form.Status
	}

	// Update image URL
	existingProject.ImageURL = imageURL

	// Update di database
	result, err := s.repository.UpdateProjekRepository(existingProject)
	if err != nil {
		// Cleanup file baru jika update gagal
		if file != nil {
			newFileName := filepath.Base(imageURL)
			newFilePath := filepath.Join(s.uploadPath, newFileName)
			os.Remove(newFilePath)
		}
		return Project{}, fmt.Errorf("gagal mengupdate projek: %v", err)
	}

	return result, nil
}

func (s *projectService) DeleteProjekService(ctx *gin.Context) error {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errors.New("ID projek tidak valid")
	}

	// Check if project exists dan ambil datanya
	existingProject, err := s.repository.GetProjekRepository(id)
	if err != nil {
		return errors.New("projek tidak ditemukan")
	}

	// Delete dari database terlebih dahulu
	err = s.repository.DeleteProjekRepository(id)
	if err != nil {
		return fmt.Errorf("gagal menghapus projek: %v", err)
	}

	// Hapus file image jika ada dan bukan default
	if existingProject.ImageURL != "" && existingProject.ImageURL != "#" {
		// Extract filename dari URL
		// ImageURL format: "/uploads/project_xxx.png"
		fileName := filepath.Base(existingProject.ImageURL)
		filePath := filepath.Join(s.uploadPath, fileName)

		// Check apakah file exists
		if _, err := os.Stat(filePath); err == nil {
			// File exists, hapus
			if err := os.Remove(filePath); err != nil {
				// Log error tapi jangan return error karena data sudah terhapus dari DB
				fmt.Printf("⚠️ Warning: gagal menghapus file %s: %v\n", filePath, err)
			} else {
				fmt.Printf("✅ File deleted successfully: %s\n", filePath)
			}
		} else {
			fmt.Printf("ℹ️ File not found, skipping deletion: %s\n", filePath)
		}
	}

	return nil
}

func (s *tagsService) CreateTags(ctx *gin.Context) (*TagResponse, error) {
	var reqcomments TagResponse
	if err := ctx.ShouldBindJSON(&reqcomments); err != nil {
		return nil, err
	}

	Tags := &ProjectTag{
		Name:  reqcomments.Name,
		Color: reqcomments.Color,
	}

	if err := s.tagsRepo.CreateTags(Tags); err != nil {
		return nil, err
	}

	return s.convertToResponse(Tags), nil

}

func (s *tagsService) convertToResponse(Tags *ProjectTag) *TagResponse {
	return &TagResponse{
		ID:        Tags.ID,
		Name:      Tags.Name,
		Color:     Tags.Color,
		CreatedAt: Tags.CreatedAt,
	}
}
