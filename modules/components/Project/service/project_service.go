package projectservice

import (
	"errors"
	"fmt"

	. "gintugas/modules/components/Project/model"
	. "gintugas/modules/components/Project/repository"
	"gintugas/modules/utils"
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
	repository    Repository
	uploadPath    string
	uploadService UploadServiceWrapper
}

// NewService untuk development (local storage)
func NewService(repository Repository, uploadPath string) Service {
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("Warning: gagal membuat folder upload: %v\n", err)
	}

	var uploadService UploadServiceWrapper

	// Cek apakah menggunakan Supabase
	if shouldUseSupabase() {
		uploadService = createSupabaseUploadService()
		if uploadService != nil {
			fmt.Println("✅ Using Supabase Storage for uploads")
		} else {
			uploadService = NewLocalUploadWrapper(
				utils.NewLocalUploadService(uploadPath),
			)
			fmt.Println("⚠️ Using Local Storage (Supabase not configured)")
		}
	} else {
		uploadService = NewLocalUploadWrapper(
			utils.NewLocalUploadService(uploadPath),
		)
		fmt.Println("ℹ️ Using Local Storage for development")
	}

	return &projectService{
		repository:    repository,
		uploadPath:    uploadPath,
		uploadService: uploadService,
	}
}

func NewServiceWithUpload(repository Repository, uploadService UploadServiceWrapper, folder string) Service {
	uploadPath := getUploadPath()
	localPath := filepath.Join(uploadPath, folder)

	if err := os.MkdirAll(localPath, 0755); err != nil {
		fmt.Printf("Warning: gagal membuat folder upload: %v\n", err)
	}

	return &projectService{
		repository:    repository,
		uploadPath:    localPath,
		uploadService: uploadService,
	}
}

func shouldUseSupabase() bool {
	uploadProvider := os.Getenv("UPLOAD_PROVIDER")
	return uploadProvider == "supabase" || os.Getenv("GIN_MODE") == "release"
}

// createSupabaseUploadService membuat service untuk upload ke Supabase
func createSupabaseUploadService() UploadServiceWrapper {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseKey == "" {
		supabaseKey = os.Getenv("SUPABASE_ANON_KEY")
	}
	bucket := os.Getenv("SUPABASE_STORAGE_BUCKET")
	if bucket == "" {
		bucket = "uploads"
	}

	if supabaseURL == "" || supabaseKey == "" {
		return nil
	}

	supabaseService := utils.NewSupabaseUploadService(supabaseURL, supabaseKey, bucket)
	return NewSupabaseUploadWrapper(supabaseService)
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

	allowedExts := []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg"}
	ext := strings.ToLower(filepath.Ext(file.Filename))

	valid := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("tipe file tidak diizinkan. File yang diizinkan: %s", strings.Join(allowedExts, ", "))
	}

	return nil
}

func (s *projectService) CreateProjekWithImageService(ctx *gin.Context) (Project, error) {
	var form ProjectForm

	// Bind form data
	if err := ctx.ShouldBind(&form); err != nil {
		fmt.Printf("❌ Bind form error: %v\n", err)
		return Project{}, fmt.Errorf("gagal binding data: %v", err)
	}

	fmt.Printf("📝 Form data received:\n")
	fmt.Printf("   Title: %s\n", form.Title)
	fmt.Printf("   Description: %s\n", form.Description)

	// Validasi required fields
	if form.Title == "" {
		return Project{}, errors.New("judul projek harus diisi")
	}

	// Handle file upload
	file, err := ctx.FormFile("image")
	imageURL := ""

	if err == nil && file != nil {
		fmt.Printf("📁 File received:\n")
		fmt.Printf("   Filename: %s\n", file.Filename)
		fmt.Printf("   Size: %d bytes\n", file.Size)
		fmt.Printf("   MIME Type: %s\n", file.Header.Get("Content-Type"))

		// Validasi file
		if err := s.validateFile(file); err != nil {
			fmt.Printf("❌ File validation failed: %v\n", err)
			return Project{}, err
		}

		fmt.Printf("🔄 Starting upload process...\n")

		// Upload file ke storage
		imageURL, err = s.uploadService.UploadFile(file, "projects")
		if err != nil {
			fmt.Printf("❌ Upload failed: %v\n", err)
			return Project{}, fmt.Errorf("gagal mengupload file: %v", err)
		}

		fmt.Printf("✅ Image uploaded successfully: %s\n", imageURL)
	} else if err != nil {
		fmt.Printf("⚠️ File error: %v\n", err)
		if err != http.ErrMissingFile {
			fmt.Printf("⚠️ Unexpected file error: %v\n", err)
		} else {
			fmt.Println("ℹ️ No image file provided, continuing without image")
		}
	}

	// Set default values
	if form.Status == "" {
		form.Status = "published"
	}
	if form.DemoURL == "" {
		form.DemoURL = "#"
	}
	if form.CodeURL == "" {
		form.CodeURL = "#"
	}

	// Convert form to Project entity
	project := Project{
		Title:        form.Title,
		Description:  form.Description,
		ImageURL:     imageURL, // URL dari Supabase atau local
		DemoURL:      form.DemoURL,
		CodeURL:      form.CodeURL,
		DisplayOrder: form.DisplayOrder,
		IsFeatured:   form.IsFeatured,
		Status:       form.Status,
	}

	fmt.Printf("💾 Saving project to database...\n")
	result, err := s.repository.CreateProjekRepository(project)
	if err != nil {
		fmt.Printf("❌ Database save failed: %v\n", err)
		// Cleanup uploaded file jika gagal menyimpan data
		if imageURL != "" {
			fmt.Printf("🧹 Cleaning up uploaded file: %s\n", imageURL)
			s.uploadService.DeleteFile(imageURL)
		}
		return Project{}, fmt.Errorf("gagal menyimpan data projek: %v", err)
	}

	fmt.Printf("✅ Project created successfully with ID: %s\n", result.ID)
	fmt.Printf("   Image URL: %s\n", result.ImageURL)

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

	// Handle file upload
	file, err := ctx.FormFile("image")

	imageURL := existingProject.ImageURL
	if err == nil && file != nil {
		fmt.Printf("✅ New file received for update: %s\n", file.Filename)

		// Validasi file
		if err := s.validateFile(file); err != nil {
			return Project{}, err
		}

		// Upload file baru
		newImageURL, err := s.uploadService.UploadFile(file, "projects")
		if err != nil {
			return Project{}, fmt.Errorf("gagal mengupload file baru: %v", err)
		}

		// Hapus file lama jika ada
		if existingProject.ImageURL != "" {
			s.uploadService.DeleteFile(existingProject.ImageURL)
		}

		imageURL = newImageURL
		fmt.Printf("🔄 Updated image to: %s\n", imageURL)
	}

	// Bind form data
	var form ProjectUpdateForm
	if err := ctx.ShouldBind(&form); err != nil {
		// Cleanup file baru jika binding gagal
		if file != nil && imageURL != existingProject.ImageURL {
			s.uploadService.DeleteFile(imageURL)
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
		if file != nil && imageURL != existingProject.ImageURL {
			s.uploadService.DeleteFile(imageURL)
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

	// Check if project exists
	existingProject, err := s.repository.GetProjekRepository(id)
	if err != nil {
		return errors.New("projek tidak ditemukan")
	}

	// Hapus file image jika ada
	if existingProject.ImageURL != "" {
		s.uploadService.DeleteFile(existingProject.ImageURL)
	}

	// Delete dari database
	return s.repository.DeleteProjekRepository(id)
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
