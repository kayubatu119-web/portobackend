package service

import (
	"errors"
	"fmt"
	model "gintugas/modules/components/all/models"
	"gintugas/modules/components/all/repo"
	"gintugas/modules/utils"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================
// UPLOAD SERVICE WRAPPER INTERFACE
// ============================

type UploadServiceWrapper interface {
	UploadFile(file *multipart.FileHeader, folder string) (string, error)
	DeleteFile(fileURL string) error
	ValidateFile(file *multipart.FileHeader, maxSizeMB int64, allowedExts []string) error
}

// SupabaseUploadWrapper
type SupabaseUploadWrapper struct {
	service *utils.SupabaseUploadService
}

func NewSupabaseUploadWrapper(service *utils.SupabaseUploadService) *SupabaseUploadWrapper {
	return &SupabaseUploadWrapper{service: service}
}

func (s *SupabaseUploadWrapper) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	return s.service.UploadFile(file, folder)
}

func (s *SupabaseUploadWrapper) DeleteFile(fileURL string) error {
	return s.service.DeleteFile(fileURL)
}

func (s *SupabaseUploadWrapper) ValidateFile(file *multipart.FileHeader, maxSizeMB int64, allowedExts []string) error {
	if file == nil {
		return errors.New("file tidak ditemukan")
	}

	// Ukuran file
	maxSize := maxSizeMB * 1024 * 1024
	if file.Size > maxSize {
		return fmt.Errorf("ukuran file maksimal %dMB", maxSizeMB)
	}

	// Extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	for _, allowed := range allowedExts {
		if ext == allowed {
			return nil
		}
	}

	return fmt.Errorf("tipe file tidak diizinkan. File yang diizinkan: %s", strings.Join(allowedExts, ", "))
}

// LocalUploadWrapper
type LocalUploadWrapper struct {
	service *utils.LocalUploadService
}

func NewLocalUploadWrapper(service *utils.LocalUploadService) *LocalUploadWrapper {
	return &LocalUploadWrapper{service: service}
}

func (l *LocalUploadWrapper) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	return l.service.UploadFile(file, folder)
}

func (l *LocalUploadWrapper) DeleteFile(fileURL string) error {
	return l.service.DeleteFile(fileURL)
}

func (l *LocalUploadWrapper) ValidateFile(file *multipart.FileHeader, maxSizeMB int64, allowedExts []string) error {
	if file == nil {
		return nil
	}

	// Simple validation untuk local
	maxSize := maxSizeMB * 1024 * 1024
	if file.Size > maxSize {
		return fmt.Errorf("ukuran file maksimal %dMB", maxSizeMB)
	}

	return nil
}

// Helper untuk menentukan upload provider
func getUploadProvider() string {
	if os.Getenv("UPLOAD_PROVIDER") == "supabase" || os.Getenv("GIN_MODE") == "release" {
		return "supabase"
	}
	return "local"
}

// ============================
// SKILLS SERVICE
// ============================

type SkillService interface {
	Create(ctx *gin.Context) (*model.SkillResponse, error)
	CreateWithIcon(ctx *gin.Context) (*model.SkillResponse, error)
	GetByID(ctx *gin.Context) (*model.SkillResponse, error)
	Update(ctx *gin.Context) (*model.SkillResponse, error)
	UpdateWithIcon(ctx *gin.Context) (*model.SkillResponse, error)
	Delete(ctx *gin.Context) error
	GetAll(ctx *gin.Context) ([]model.SkillResponse, error)
	GetFeatured(ctx *gin.Context) ([]model.SkillResponse, error)
	GetByCategory(ctx *gin.Context) ([]model.SkillResponse, error)
}

type skillService struct {
	repo          repo.SkillRepository
	uploadPath    string
	uploadService UploadServiceWrapper
}

// NewSkillService untuk local storage
func NewSkillService(repo repo.SkillRepository, uploadPath string) SkillService {
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("⚠️ Warning: gagal membuat folder upload skill: %v\n", err)
	}

	var uploadService UploadServiceWrapper

	if getUploadProvider() == "supabase" {
		supabaseService := createSupabaseUploadService()
		if supabaseService != nil {
			uploadService = NewSupabaseUploadWrapper(supabaseService)
			fmt.Println("✅ Using Supabase Storage for skills")
		} else {
			localService := utils.NewLocalUploadService(uploadPath)
			uploadService = NewLocalUploadWrapper(localService)
			fmt.Println("⚠️ Using Local Storage for skills (Supabase not configured)")
		}
	} else {
		localService := utils.NewLocalUploadService(uploadPath)
		uploadService = NewLocalUploadWrapper(localService)
		fmt.Println("ℹ️ Using Local Storage for skills (development)")
	}

	return &skillService{
		repo:          repo,
		uploadPath:    uploadPath,
		uploadService: uploadService,
	}
}

// NewSkillServiceWithUpload untuk custom upload service
func NewSkillServiceWithUpload(repo repo.SkillRepository, uploadService UploadServiceWrapper, folder string) SkillService {
	uploadPath := getUploadPath()
	localPath := filepath.Join(uploadPath, folder)
	if err := os.MkdirAll(localPath, 0755); err != nil {
		fmt.Printf("⚠️ Warning: gagal membuat folder upload: %v\n", err)
	}

	return &skillService{
		repo:          repo,
		uploadPath:    localPath,
		uploadService: uploadService,
	}
}

// Helper functions
func createSupabaseUploadService() *utils.SupabaseUploadService {
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

	return utils.NewSupabaseUploadService(supabaseURL, supabaseKey, bucket)
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

func (s *skillService) Create(ctx *gin.Context) (*model.SkillResponse, error) {
	var req model.SkillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	skill := &model.Skill{
		Name:         req.Name,
		Value:        req.Value,
		IconURL:      req.IconURL,
		Category:     req.Category,
		DisplayOrder: req.DisplayOrder,
		IsFeatured:   req.IsFeatured,
	}

	if err := s.repo.Create(skill); err != nil {
		return nil, err
	}

	return s.convertSkillToResponse(skill), nil
}

func (s *skillService) CreateWithIcon(ctx *gin.Context) (*model.SkillResponse, error) {
	var form model.SkillForm

	// Bind form data
	if err := ctx.ShouldBind(&form); err != nil {
		return nil, fmt.Errorf("gagal binding data: %v", err)
	}

	// Validasi required fields
	if form.Name == "" {
		return nil, errors.New("nama skill harus diisi")
	}

	if form.Value < 0 || form.Value > 100 {
		return nil, errors.New("nilai skill harus antara 0-100")
	}

	// Handle file upload
	file, err := ctx.FormFile("icon")
	if err != nil && err != http.ErrMissingFile {
		return nil, fmt.Errorf("gagal mengambil file icon: %v", err)
	}

	iconURL := ""
	if file != nil {
		// Validasi file
		allowedExts := []string{".jpg", ".jpeg", ".png", ".webp", ".svg", ".ico"}
		if err := s.uploadService.ValidateFile(file, 5, allowedExts); err != nil {
			return nil, err
		}

		// Upload ke Supabase atau Local storage
		iconURL, err = s.uploadService.UploadFile(file, "skills")
		if err != nil {
			return nil, fmt.Errorf("gagal upload file icon: %v", err)
		}
		fmt.Printf("✅ Skill icon uploaded: %s\n", iconURL)
	}

	// Set default values
	if form.Category == "" {
		form.Category = "programming"
	}
	if form.DisplayOrder == 0 {
		form.DisplayOrder = 0
	}

	// Create skill entity
	skill := &model.Skill{
		Name:         form.Name,
		Value:        form.Value,
		IconURL:      iconURL,
		Category:     form.Category,
		DisplayOrder: form.DisplayOrder,
		IsFeatured:   form.IsFeatured,
	}

	// Save to database
	if err := s.repo.Create(skill); err != nil {
		// Cleanup file jika gagal save ke database
		if file != nil && iconURL != "" {
			s.uploadService.DeleteFile(iconURL)
		}
		return nil, fmt.Errorf("gagal menyimpan data skill: %v", err)
	}

	return s.convertSkillToResponse(skill), nil
}

func (s *skillService) GetByID(ctx *gin.Context) (*model.SkillResponse, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, errors.New("invalid skill ID")
	}

	skill, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.convertSkillToResponse(skill), nil
}

func (s *skillService) Update(ctx *gin.Context) (*model.SkillResponse, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, errors.New("invalid skill ID")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	var req model.SkillUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Value != 0 {
		existing.Value = req.Value
	}
	existing.IconURL = req.IconURL
	existing.Category = req.Category
	existing.DisplayOrder = req.DisplayOrder
	existing.IsFeatured = req.IsFeatured
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	return s.convertSkillToResponse(existing), nil
}

func (s *skillService) UpdateWithIcon(ctx *gin.Context) (*model.SkillResponse, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, errors.New("invalid skill ID")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	var form model.SkillForm
	if err := ctx.ShouldBind(&form); err != nil {
		return nil, fmt.Errorf("gagal binding data: %v", err)
	}

	// Handle file upload
	file, err := ctx.FormFile("icon")
	if err != nil && err != http.ErrMissingFile {
		return nil, fmt.Errorf("gagal mengambil file icon: %v", err)
	}

	// Jika ada file baru diupload
	if file != nil {
		// Validasi file
		allowedExts := []string{".jpg", ".jpeg", ".png", ".webp", ".svg", ".ico"}
		if err := s.uploadService.ValidateFile(file, 5, allowedExts); err != nil {
			return nil, err
		}

		// Upload file baru
		newIconURL, err := s.uploadService.UploadFile(file, "skills")
		if err != nil {
			return nil, fmt.Errorf("gagal upload file icon: %v", err)
		}

		// Hapus file lama jika ada
		if existing.IconURL != "" {
			s.uploadService.DeleteFile(existing.IconURL)
		}

		// Update icon URL dengan yang baru
		existing.IconURL = newIconURL
	}

	// Update fields lainnya
	if form.Name != "" {
		existing.Name = form.Name
	}
	if form.Value != 0 {
		existing.Value = form.Value
	}
	if form.Category != "" {
		existing.Category = form.Category
	}
	existing.DisplayOrder = form.DisplayOrder
	existing.IsFeatured = form.IsFeatured
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(existing); err != nil {
		// Cleanup file baru jika gagal update
		if file != nil && existing.IconURL != "" {
			s.uploadService.DeleteFile(existing.IconURL)
		}
		return nil, fmt.Errorf("gagal mengupdate data skill: %v", err)
	}

	return s.convertSkillToResponse(existing), nil
}

func (s *skillService) Delete(ctx *gin.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return errors.New("invalid skill ID")
	}

	// Get skill data untuk menghapus file icon
	skill, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Hapus file icon dari Supabase atau local storage jika ada
	if skill.IconURL != "" {
		if err := s.uploadService.DeleteFile(skill.IconURL); err != nil {
			fmt.Printf("⚠️ Warning: gagal hapus file icon: %v\n", err)
			// Jangan return error, lanjut delete dari DB
		}
	}

	return s.repo.Delete(id)
}

func (s *skillService) GetAll(ctx *gin.Context) ([]model.SkillResponse, error) {
	skills, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []model.SkillResponse
	for _, skill := range skills {
		responses = append(responses, *s.convertSkillToResponse(&skill))
	}

	return responses, nil
}

func (s *skillService) GetFeatured(ctx *gin.Context) ([]model.SkillResponse, error) {
	skills, err := s.repo.GetFeatured()
	if err != nil {
		return nil, err
	}

	var responses []model.SkillResponse
	for _, skill := range skills {
		responses = append(responses, *s.convertSkillToResponse(&skill))
	}

	return responses, nil
}

func (s *skillService) GetByCategory(ctx *gin.Context) ([]model.SkillResponse, error) {
	category := ctx.Param("category")
	skills, err := s.repo.GetByCategory(category)
	if err != nil {
		return nil, err
	}

	var responses []model.SkillResponse
	for _, skill := range skills {
		responses = append(responses, *s.convertSkillToResponse(&skill))
	}

	return responses, nil
}

func (s *skillService) convertSkillToResponse(skill *model.Skill) *model.SkillResponse {
	return &model.SkillResponse{
		ID:           skill.ID,
		Name:         skill.Name,
		Value:        skill.Value,
		IconURL:      skill.IconURL,
		Category:     skill.Category,
		DisplayOrder: skill.DisplayOrder,
		IsFeatured:   skill.IsFeatured,
		CreatedAt:    skill.CreatedAt,
		UpdatedAt:    skill.UpdatedAt,
	}
}

// ============================
// CERTIFICATES SERVICE
// ============================

type CertificateService interface {
	Create(ctx *gin.Context) (*model.CertificateResponse, error)
	CreateWithImage(ctx *gin.Context) (*model.CertificateResponse, error)
	GetByID(ctx *gin.Context) (*model.CertificateResponse, error)
	Update(ctx *gin.Context) (*model.CertificateResponse, error)
	Delete(ctx *gin.Context) error
	GetAll(ctx *gin.Context) ([]model.CertificateResponse, error)
}

type certificateService struct {
	repo          repo.CertificateRepository
	uploadPath    string
	uploadService UploadServiceWrapper
}

// NewCertificateService untuk local storage
func NewCertificateService(repo repo.CertificateRepository, uploadPath string) CertificateService {
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("⚠️ Warning: gagal membuat folder upload certificate: %v\n", err)
	}

	var uploadService UploadServiceWrapper

	if getUploadProvider() == "supabase" {
		supabaseService := createSupabaseUploadService()
		if supabaseService != nil {
			uploadService = NewSupabaseUploadWrapper(supabaseService)
			fmt.Println("✅ Using Supabase Storage for certificates")
		} else {
			localService := utils.NewLocalUploadService(uploadPath)
			uploadService = NewLocalUploadWrapper(localService)
			fmt.Println("⚠️ Using Local Storage for certificates (Supabase not configured)")
		}
	} else {
		localService := utils.NewLocalUploadService(uploadPath)
		uploadService = NewLocalUploadWrapper(localService)
		fmt.Println("ℹ️ Using Local Storage for certificates (development)")
	}

	return &certificateService{
		repo:          repo,
		uploadPath:    uploadPath,
		uploadService: uploadService,
	}
}

// NewCertificateServiceWithUpload untuk custom upload service
func NewCertificateServiceWithUpload(repo repo.CertificateRepository, uploadService UploadServiceWrapper, folder string) CertificateService {
	uploadPath := getUploadPath()
	localPath := filepath.Join(uploadPath, folder)
	if err := os.MkdirAll(localPath, 0755); err != nil {
		fmt.Printf("⚠️ Warning: gagal membuat folder upload: %v\n", err)
	}

	return &certificateService{
		repo:          repo,
		uploadPath:    localPath,
		uploadService: uploadService,
	}
}

func (s *certificateService) Create(ctx *gin.Context) (*model.CertificateResponse, error) {
	var req model.CertificateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	cert := &model.Certificate{
		Name:          req.Name,
		ImageURL:      req.ImageURL,
		IssueDate:     req.IssueDate,
		Issuer:        req.Issuer,
		CredentialURL: req.CredentialURL,
		DisplayOrder:  req.DisplayOrder,
	}

	if err := s.repo.Create(cert); err != nil {
		return nil, err
	}

	return s.convertCertToResponse(cert), nil
}

func (s *certificateService) CreateWithImage(ctx *gin.Context) (*model.CertificateResponse, error) {
	var form model.CertificateForm

	// Bind form data
	if err := ctx.ShouldBind(&form); err != nil {
		return nil, fmt.Errorf("gagal binding data: %v", err)
	}

	// Validasi required fields
	if form.Name == "" {
		return nil, errors.New("nama sertifikat harus diisi")
	}

	// Handle file upload
	file, err := ctx.FormFile("image")
	if err != nil {
		return nil, fmt.Errorf("file gambar harus diupload: %v", err)
	}

	// Validasi file
	allowedExts := []string{".jpg", ".jpeg", ".png", ".webp", ".pdf"}
	if err := s.uploadService.ValidateFile(file, 10, allowedExts); err != nil {
		return nil, err
	}

	// Upload ke Supabase atau Local storage
	imageURL, err := s.uploadService.UploadFile(file, "certificates")
	if err != nil {
		return nil, fmt.Errorf("gagal upload file: %v", err)
	}
	fmt.Printf("✅ Certificate image uploaded: %s\n", imageURL)

	// Parse issue date
	var issueDate time.Time
	if form.IssueDate != "" {
		parsedDate, err := time.Parse("2006-01-02", form.IssueDate)
		if err != nil {
			// Cleanup file jika parsing gagal
			s.uploadService.DeleteFile(imageURL)
			return nil, fmt.Errorf("format tanggal tidak valid, gunakan format YYYY-MM-DD: %v", err)
		}
		issueDate = parsedDate
	}

	// Set default values
	if form.DisplayOrder == 0 {
		form.DisplayOrder = 0
	}
	if form.Issuer == "" {
		form.Issuer = "-"
	}

	// Create certificate entity
	cert := &model.Certificate{
		Name:          form.Name,
		ImageURL:      imageURL, // Full URL dari Supabase atau local
		IssueDate:     issueDate,
		Issuer:        form.Issuer,
		CredentialURL: form.CredentialURL,
		DisplayOrder:  form.DisplayOrder,
	}

	// Save to database
	if err := s.repo.Create(cert); err != nil {
		// Cleanup file jika gagal save ke database
		s.uploadService.DeleteFile(imageURL)
		return nil, fmt.Errorf("gagal menyimpan data sertifikat: %v", err)
	}

	return s.convertCertToResponse(cert), nil
}

func (s *certificateService) GetByID(ctx *gin.Context) (*model.CertificateResponse, error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, errors.New("ID sertifikat tidak valid")
	}

	cert, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.convertCertToResponse(cert), nil
}

func (s *certificateService) Update(ctx *gin.Context) (*model.CertificateResponse, error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, errors.New("ID sertifikat tidak valid")
	}

	// Check if certificate exists
	existingCert, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("sertifikat tidak ditemukan")
	}

	var updateData model.CertificateUpdateRequest
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		return nil, err
	}

	// Update fields
	if updateData.Name != "" {
		existingCert.Name = updateData.Name
	}
	if updateData.ImageURL != "" {
		existingCert.ImageURL = updateData.ImageURL
	}
	if !updateData.IssueDate.IsZero() {
		existingCert.IssueDate = updateData.IssueDate
	}
	if updateData.Issuer != "" {
		existingCert.Issuer = updateData.Issuer
	}
	if updateData.CredentialURL != "" {
		existingCert.CredentialURL = updateData.CredentialURL
	}
	existingCert.DisplayOrder = updateData.DisplayOrder

	if err := s.repo.Update(existingCert); err != nil {
		return nil, err
	}

	return s.convertCertToResponse(existingCert), nil
}

func (s *certificateService) Delete(ctx *gin.Context) error {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errors.New("ID sertifikat tidak valid")
	}

	// Get certificate data untuk hapus file
	cert, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("sertifikat tidak ditemukan")
	}

	// Delete dari database terlebih dahulu
	err = s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("gagal menghapus sertifikat: %v", err)
	}

	// Hapus file image dari Supabase atau local storage jika ada
	if cert.ImageURL != "" && cert.ImageURL != "#" {
		if err := s.uploadService.DeleteFile(cert.ImageURL); err != nil {
			fmt.Printf("⚠️ Warning: gagal menghapus file image: %v\n", err)
			// Jangan return error karena data sudah terhapus dari DB
		}
	}

	return nil
}

func (s *certificateService) GetAll(ctx *gin.Context) ([]model.CertificateResponse, error) {
	certs, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := make([]model.CertificateResponse, len(certs))
	for i, cert := range certs {
		responses[i] = *s.convertCertToResponse(&cert)
	}

	return responses, nil
}

func (s *certificateService) convertCertToResponse(cert *model.Certificate) *model.CertificateResponse {
	return &model.CertificateResponse{
		ID:            cert.ID,
		Name:          cert.Name,
		ImageURL:      cert.ImageURL,
		IssueDate:     cert.IssueDate,
		Issuer:        cert.Issuer,
		CredentialURL: cert.CredentialURL,
		DisplayOrder:  cert.DisplayOrder,
		CreatedAt:     cert.CreatedAt,
	}
}

// ============================
// EDUCATION SERVICE (no upload needed)
// ============================

type EducationService interface {
	CreateWithAchievements(ctx *gin.Context) (*model.EducationResponse, error)
	GetByIDWithAchievements(ctx *gin.Context) (*model.EducationResponse, error)
	UpdateWithAchievements(ctx *gin.Context) (*model.EducationResponse, error)
	DeleteWithAchievements(ctx *gin.Context) error
	GetAllWithAchievements(ctx *gin.Context) ([]model.EducationResponse, error)
}

type educationService struct {
	repo repo.EducationRepository
}

func NewEducationService(repo repo.EducationRepository) EducationService {
	return &educationService{repo: repo}
}

func (s *educationService) CreateWithAchievements(ctx *gin.Context) (*model.EducationResponse, error) {
	var req model.EducationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	edu := &model.Education{
		School:       req.School,
		Major:        req.Major,
		StartYear:    req.StartYear,
		EndYear:      req.EndYear,
		Description:  req.Description,
		Degree:       req.Degree,
		DisplayOrder: req.DisplayOrder,
	}

	for _, achReq := range req.Achievements {
		edu.Achievements = append(edu.Achievements, model.EducationAchievement{
			Achievement:  achReq.Achievement,
			DisplayOrder: achReq.DisplayOrder,
		})
	}

	if err := s.repo.CreateWithAchievements(edu); err != nil {
		return nil, err
	}

	return convertEducationToResponse(edu), nil
}

func (s *educationService) GetByIDWithAchievements(ctx *gin.Context) (*model.EducationResponse, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, errors.New("invalid education ID")
	}

	edu, err := s.repo.GetByIDWithAchievements(id)
	if err != nil {
		return nil, err
	}

	return convertEducationToResponse(edu), nil
}

func (s *educationService) UpdateWithAchievements(ctx *gin.Context) (*model.EducationResponse, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, errors.New("invalid education ID")
	}

	existing, err := s.repo.GetByIDWithAchievements(id)
	if err != nil {
		return nil, err
	}

	var req model.EducationUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	if req.School != "" {
		existing.School = req.School
	}
	if req.Major != "" {
		existing.Major = req.Major
	}
	existing.StartYear = req.StartYear
	existing.EndYear = req.EndYear
	existing.Description = req.Description
	existing.Degree = req.Degree
	existing.DisplayOrder = req.DisplayOrder
	existing.UpdatedAt = time.Now()

	existing.Achievements = nil
	for _, achReq := range req.Achievements {
		existing.Achievements = append(existing.Achievements, model.EducationAchievement{
			Achievement:  achReq.Achievement,
			DisplayOrder: achReq.DisplayOrder,
		})
	}

	if err := s.repo.UpdateWithAchievements(existing); err != nil {
		return nil, err
	}

	return convertEducationToResponse(existing), nil
}

func (s *educationService) DeleteWithAchievements(ctx *gin.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return errors.New("invalid education ID")
	}

	return s.repo.DeleteWithAchievements(id)
}

func (s *educationService) GetAllWithAchievements(ctx *gin.Context) ([]model.EducationResponse, error) {
	educations, err := s.repo.GetAllWithAchievements()
	if err != nil {
		return nil, err
	}

	var responses []model.EducationResponse
	for _, edu := range educations {
		responses = append(responses, *convertEducationToResponse(&edu))
	}

	return responses, nil
}

// ============================
// TESTIMONIALS SERVICE (no upload needed)
// ============================

type TestimonialService interface {
	Create(ctx *gin.Context) (*model.TestimonialResponse, error)
	GetByID(ctx *gin.Context) (*model.TestimonialResponse, error)
	Update(ctx *gin.Context) (*model.TestimonialResponse, error)
	Delete(ctx *gin.Context) error
	GetAll(ctx *gin.Context) ([]model.TestimonialResponse, error)
	GetFeatured(ctx *gin.Context) ([]model.TestimonialResponse, error)
	GetByStatus(ctx *gin.Context) ([]model.TestimonialResponse, error)
}

type testimonialService struct {
	repo repo.TestimonialRepository
}

func NewTestimonialService(repo repo.TestimonialRepository) TestimonialService {
	return &testimonialService{repo: repo}
}

func (s *testimonialService) Create(ctx *gin.Context) (*model.TestimonialResponse, error) {
	var req model.TestimonialRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	test := &model.Testimonial{
		Name:         req.Name,
		Title:        req.Title,
		Message:      req.Message,
		AvatarURL:    req.AvatarURL,
		Rating:       req.Rating,
		IsFeatured:   req.IsFeatured,
		DisplayOrder: req.DisplayOrder,
		Status:       req.Status,
	}

	if test.Status == "" {
		test.Status = "approved"
	}

	if err := s.repo.Create(test); err != nil {
		return nil, err
	}

	return convertTestimonialToResponse(test), nil
}

func (s *testimonialService) GetByID(ctx *gin.Context) (*model.TestimonialResponse, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, errors.New("invalid testimonial ID")
	}

	test, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return convertTestimonialToResponse(test), nil
}

func (s *testimonialService) Update(ctx *gin.Context) (*model.TestimonialResponse, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, errors.New("invalid testimonial ID")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	var req model.TestimonialUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Message != "" {
		existing.Message = req.Message
	}
	existing.AvatarURL = req.AvatarURL
	if req.Rating != 0 {
		existing.Rating = req.Rating
	}
	existing.IsFeatured = req.IsFeatured
	existing.DisplayOrder = req.DisplayOrder
	if req.Status != "" {
		existing.Status = req.Status
	}

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	return convertTestimonialToResponse(existing), nil
}

func (s *testimonialService) Delete(ctx *gin.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return errors.New("invalid testimonial ID")
	}

	return s.repo.Delete(id)
}

func (s *testimonialService) GetAll(ctx *gin.Context) ([]model.TestimonialResponse, error) {
	testimonials, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []model.TestimonialResponse
	for _, test := range testimonials {
		responses = append(responses, *convertTestimonialToResponse(&test))
	}

	return responses, nil
}

func (s *testimonialService) GetFeatured(ctx *gin.Context) ([]model.TestimonialResponse, error) {
	testimonials, err := s.repo.GetFeatured()
	if err != nil {
		return nil, err
	}

	var responses []model.TestimonialResponse
	for _, test := range testimonials {
		responses = append(responses, *convertTestimonialToResponse(&test))
	}

	return responses, nil
}

func (s *testimonialService) GetByStatus(ctx *gin.Context) ([]model.TestimonialResponse, error) {
	status := ctx.Param("status")
	testimonials, err := s.repo.GetByStatus(status)
	if err != nil {
		return nil, err
	}

	var responses []model.TestimonialResponse
	for _, test := range testimonials {
		responses = append(responses, *convertTestimonialToResponse(&test))
	}

	return responses, nil
}

// ============================
// BLOG SERVICE (no upload needed)
// ============================

type BlogService interface {
	CreateWithTags(ctx *gin.Context) (*model.BlogPostResponse, error)
	GetByIDWithTags(ctx *gin.Context) (*model.BlogPostResponse, error)
	GetBySlugWithTags(ctx *gin.Context) (*model.BlogPostResponse, error)
	UpdateWithTags(ctx *gin.Context) (*model.BlogPostResponse, error)
	DeleteWithTags(ctx *gin.Context) error
	GetAllWithTags(ctx *gin.Context) ([]model.BlogPostResponse, error)
	GetPublishedWithTags(ctx *gin.Context) ([]model.BlogPostResponse, error)
	GetAllTags(ctx *gin.Context) ([]model.TagResponse, error)
}

type blogService struct {
	repo repo.BlogRepository
}

func NewBlogService(repo repo.BlogRepository) BlogService {
	return &blogService{repo: repo}
}

func (s *blogService) CreateWithTags(ctx *gin.Context) (*model.BlogPostResponse, error) {
	var req model.BlogPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	post := &model.BlogPost{
		Title:         req.Title,
		Content:       req.Content,
		Excerpt:       req.Excerpt,
		Slug:          req.Slug,
		FeaturedImage: req.FeaturedImage,
		PublishDate:   req.PublishDate,
		Status:        req.Status,
	}

	if post.Status == "" {
		post.Status = "draft"
	}

	for _, tagReq := range req.Tags {
		post.Tags = append(post.Tags, model.BlogTag{Name: tagReq.Name})
	}

	if err := s.repo.CreateWithTags(post); err != nil {
		return nil, err
	}

	return convertBlogToResponse(post), nil
}

func (s *blogService) GetByIDWithTags(ctx *gin.Context) (*model.BlogPostResponse, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, errors.New("invalid post ID")
	}

	post, err := s.repo.GetByIDWithTags(id)
	if err != nil {
		return nil, err
	}

	// Increment view count
	_ = s.repo.IncrementViewCount(id)

	return convertBlogToResponse(post), nil
}

func (s *blogService) GetBySlugWithTags(ctx *gin.Context) (*model.BlogPostResponse, error) {
	slug := ctx.Param("slug")

	post, err := s.repo.GetBySlugWithTags(slug)
	if err != nil {
		return nil, err
	}

	// Increment view count
	_ = s.repo.IncrementViewCount(post.ID)

	return convertBlogToResponse(post), nil
}

func (s *blogService) UpdateWithTags(ctx *gin.Context) (*model.BlogPostResponse, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, errors.New("invalid post ID")
	}

	existing, err := s.repo.GetByIDWithTags(id)
	if err != nil {
		return nil, err
	}

	var req model.BlogPostUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	existing.Content = req.Content
	existing.Excerpt = req.Excerpt
	if req.Slug != "" {
		existing.Slug = req.Slug
	}
	existing.FeaturedImage = req.FeaturedImage
	if !req.PublishDate.IsZero() {
		existing.PublishDate = req.PublishDate
	}
	if req.Status != "" {
		existing.Status = req.Status
	}
	existing.UpdatedAt = time.Now()

	existing.Tags = nil
	for _, tagReq := range req.Tags {
		existing.Tags = append(existing.Tags, model.BlogTag{Name: tagReq.Name})
	}

	if err := s.repo.UpdateWithTags(existing); err != nil {
		return nil, err
	}

	return convertBlogToResponse(existing), nil
}

func (s *blogService) DeleteWithTags(ctx *gin.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return errors.New("invalid post ID")
	}

	return s.repo.DeleteWithTags(id)
}

func (s *blogService) GetAllWithTags(ctx *gin.Context) ([]model.BlogPostResponse, error) {
	posts, err := s.repo.GetAllWithTags()
	if err != nil {
		return nil, err
	}

	var responses []model.BlogPostResponse
	for _, post := range posts {
		responses = append(responses, *convertBlogToResponse(&post))
	}

	return responses, nil
}

func (s *blogService) GetPublishedWithTags(ctx *gin.Context) ([]model.BlogPostResponse, error) {
	posts, err := s.repo.GetPublishedWithTags()
	if err != nil {
		return nil, err
	}

	var responses []model.BlogPostResponse
	for _, post := range posts {
		responses = append(responses, *convertBlogToResponse(&post))
	}

	return responses, nil
}

func (s *blogService) GetAllTags(ctx *gin.Context) ([]model.TagResponse, error) {
	tags, err := s.repo.GetAllTags()
	if err != nil {
		return nil, err
	}

	var responses []model.TagResponse
	for _, tag := range tags {
		responses = append(responses, model.TagResponse{
			ID:        tag.ID,
			Name:      tag.Name,
			CreatedAt: tag.CreatedAt,
		})
	}

	return responses, nil
}

// ============================
// SECTIONS SERVICE (no upload needed)
// ============================

type SectionService interface {
	Create(ctx *gin.Context) (*model.SectionResponse, error)
	Delete(ctx *gin.Context) error
	GetAll(ctx *gin.Context) ([]model.SectionResponse, error)
}

type sectionService struct {
	repo repo.SectionRepository
}

func NewSectionService(repo repo.SectionRepository) SectionService {
	return &sectionService{repo: repo}
}

func (s *sectionService) Create(ctx *gin.Context) (*model.SectionResponse, error) {
	var req model.SectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	section := &model.Section{
		SectionID:    req.SectionID,
		Label:        req.Label,
		DisplayOrder: req.DisplayOrder,
		IsActive:     req.IsActive,
	}

	if err := s.repo.Create(section); err != nil {
		return nil, err
	}

	return convertSectionToResponse(section), nil
}

func (s *sectionService) Delete(ctx *gin.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return errors.New("invalid section ID")
	}

	return s.repo.Delete(id)
}

func (s *sectionService) GetAll(ctx *gin.Context) ([]model.SectionResponse, error) {
	sections, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []model.SectionResponse
	for _, section := range sections {
		responses = append(responses, *convertSectionToResponse(&section))
	}

	return responses, nil
}

// ============================
// SOCIAL LINKS SERVICE (no upload needed)
// ============================

type SocialLinkService interface {
	Create(ctx *gin.Context) (*model.SocialLinkResponse, error)
	Delete(ctx *gin.Context) error
	GetAll(ctx *gin.Context) ([]model.SocialLinkResponse, error)
}

type socialLinkService struct {
	repo repo.SocialLinkRepository
}

func NewSocialLinkService(repo repo.SocialLinkRepository) SocialLinkService {
	return &socialLinkService{repo: repo}
}

func (s *socialLinkService) Create(ctx *gin.Context) (*model.SocialLinkResponse, error) {
	var req model.SocialLinkRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	link := &model.SocialLink{
		Platform:     req.Platform,
		URL:          req.URL,
		IconName:     req.IconName,
		DisplayOrder: req.DisplayOrder,
		IsActive:     req.IsActive,
	}

	if err := s.repo.Create(link); err != nil {
		return nil, err
	}

	return convertSocialLinkToResponse(link), nil
}

func (s *socialLinkService) Delete(ctx *gin.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return errors.New("invalid social link ID")
	}

	return s.repo.Delete(id)
}

func (s *socialLinkService) GetAll(ctx *gin.Context) ([]model.SocialLinkResponse, error) {
	links, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []model.SocialLinkResponse
	for _, link := range links {
		responses = append(responses, *convertSocialLinkToResponse(&link))
	}

	return responses, nil
}

// ============================
// SETTINGS SERVICE (no upload needed)
// ============================

type SettingService interface {
	Create(ctx *gin.Context) (*model.SettingResponse, error)
	Delete(ctx *gin.Context) error
	GetAll(ctx *gin.Context) ([]model.SettingResponse, error)
}

type settingService struct {
	repo repo.SettingRepository
}

func NewSettingService(repo repo.SettingRepository) SettingService {
	return &settingService{repo: repo}
}

func (s *settingService) Create(ctx *gin.Context) (*model.SettingResponse, error) {
	var req model.SettingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	setting := &model.Setting{
		Key:         req.Key,
		Value:       req.Value,
		DataType:    req.DataType,
		Description: req.Description,
	}

	if setting.DataType == "" {
		setting.DataType = "string"
	}

	if err := s.repo.Create(setting); err != nil {
		return nil, err
	}

	return convertSettingToResponse(setting), nil
}

func (s *settingService) Delete(ctx *gin.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return errors.New("invalid setting ID")
	}

	return s.repo.Delete(id)
}

func (s *settingService) GetAll(ctx *gin.Context) ([]model.SettingResponse, error) {
	settings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []model.SettingResponse
	for _, setting := range settings {
		responses = append(responses, *convertSettingToResponse(&setting))
	}

	return responses, nil
}

// ============================
// HELPER FUNCTIONS
// ============================

func convertEducationToResponse(edu *model.Education) *model.EducationResponse {
	var achievements []model.AchievementResponse
	for _, ach := range edu.Achievements {
		achievements = append(achievements, model.AchievementResponse{
			ID:           ach.ID,
			EducationID:  ach.EducationID,
			Achievement:  ach.Achievement,
			DisplayOrder: ach.DisplayOrder,
			CreatedAt:    ach.CreatedAt,
		})
	}

	return &model.EducationResponse{
		ID:           edu.ID,
		School:       edu.School,
		Major:        edu.Major,
		StartYear:    edu.StartYear,
		EndYear:      edu.EndYear,
		Description:  edu.Description,
		Degree:       edu.Degree,
		DisplayOrder: edu.DisplayOrder,
		Achievements: achievements,
		CreatedAt:    edu.CreatedAt,
		UpdatedAt:    edu.UpdatedAt,
	}
}

func convertTestimonialToResponse(test *model.Testimonial) *model.TestimonialResponse {
	return &model.TestimonialResponse{
		ID:           test.ID,
		Name:         test.Name,
		Title:        test.Title,
		Message:      test.Message,
		AvatarURL:    test.AvatarURL,
		Rating:       test.Rating,
		IsFeatured:   test.IsFeatured,
		DisplayOrder: test.DisplayOrder,
		Status:       test.Status,
		CreatedAt:    test.CreatedAt,
	}
}

func convertBlogToResponse(post *model.BlogPost) *model.BlogPostResponse {
	var tags []model.TagResponse
	for _, tag := range post.Tags {
		tags = append(tags, model.TagResponse{
			ID:        tag.ID,
			Name:      tag.Name,
			CreatedAt: tag.CreatedAt,
		})
	}

	return &model.BlogPostResponse{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		Excerpt:       post.Excerpt,
		Slug:          post.Slug,
		FeaturedImage: post.FeaturedImage,
		PublishDate:   post.PublishDate,
		Status:        post.Status,
		ViewCount:     post.ViewCount,
		Tags:          tags,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
	}
}

func convertSectionToResponse(section *model.Section) *model.SectionResponse {
	return &model.SectionResponse{
		ID:           section.ID,
		SectionID:    section.SectionID,
		Label:        section.Label,
		DisplayOrder: section.DisplayOrder,
		IsActive:     section.IsActive,
		CreatedAt:    section.CreatedAt,
		UpdatedAt:    section.UpdatedAt,
	}
}

func convertSocialLinkToResponse(link *model.SocialLink) *model.SocialLinkResponse {
	return &model.SocialLinkResponse{
		ID:           link.ID,
		Platform:     link.Platform,
		URL:          link.URL,
		IconName:     link.IconName,
		DisplayOrder: link.DisplayOrder,
		IsActive:     link.IsActive,
		CreatedAt:    link.CreatedAt,
		UpdatedAt:    link.UpdatedAt,
	}
}

func convertSettingToResponse(setting *model.Setting) *model.SettingResponse {
	return &model.SettingResponse{
		ID:          setting.ID,
		Key:         setting.Key,
		Value:       setting.Value,
		DataType:    setting.DataType,
		Description: setting.Description,
		CreatedAt:   setting.CreatedAt,
		UpdatedAt:   setting.UpdatedAt,
	}
}
