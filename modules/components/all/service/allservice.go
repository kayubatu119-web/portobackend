package service

import (
	"errors"
	"fmt"
	model "gintugas/modules/components/all/models"
	"gintugas/modules/components/all/repo"
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
// SKILLS SERVICE
// ============================

type SkillService interface {
	Create(ctx *gin.Context) (*model.SkillResponse, error)
	CreateWithIcon(ctx *gin.Context) (*model.SkillResponse, error) // Tambah method baru
	GetByID(ctx *gin.Context) (*model.SkillResponse, error)
	Update(ctx *gin.Context) (*model.SkillResponse, error)
	UpdateWithIcon(ctx *gin.Context) (*model.SkillResponse, error) // Tambah method baru
	Delete(ctx *gin.Context) error
	GetAll(ctx *gin.Context) ([]model.SkillResponse, error)
	GetFeatured(ctx *gin.Context) ([]model.SkillResponse, error)
	GetByCategory(ctx *gin.Context) ([]model.SkillResponse, error)
}

type skillService struct {
	repo       repo.SkillRepository
	uploadPath string
}

func NewSkillService(repo repo.SkillRepository, uploadPath string) SkillService {
	// Buat folder upload jika belum ada
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("Warning: gagal membuat folder upload skill: %v\n", err)
	}
	return &skillService{
		repo:       repo,
		uploadPath: uploadPath,
	}
}

func (s *skillService) validateFile(file *multipart.FileHeader) error {
	// Ukuran file 5MB (untuk icon biasanya lebih kecil)
	maxSize := int64(5 * 1024 * 1024)
	if file.Size > maxSize {
		return errors.New("ukuran file maksimal 5MB")
	}

	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".svg":  true, // SVG bagus untuk icon
		".ico":  true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		return errors.New("tipe file tidak diizinkan. File yang diizinkan: jpg, jpeg, png, webp, svg, ico")
	}

	return nil
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
		if err := s.validateFile(file); err != nil {
			return nil, err
		}

		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("skill_%s%s", uuid.New().String(), ext)
		filePath := filepath.Join(s.uploadPath, fileName)

		// Simpan file
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			return nil, fmt.Errorf("gagal menyimpan file icon: %v", err)
		}

		iconURL = "/uploads/skills/" + fileName
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
			os.Remove(filepath.Join(s.uploadPath, filepath.Base(iconURL)))
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
		if err := s.validateFile(file); err != nil {
			return nil, err
		}

		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("skill_%s%s", uuid.New().String(), ext)
		filePath := filepath.Join(s.uploadPath, fileName)

		// Simpan file baru
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			return nil, fmt.Errorf("gagal menyimpan file icon: %v", err)
		}

		// Hapus file lama jika ada
		if existing.IconURL != "" {
			oldFileName := filepath.Base(existing.IconURL)
			oldFilePath := filepath.Join(s.uploadPath, oldFileName)
			os.Remove(oldFilePath) // Ignore error jika file tidak ada
		}

		// Update icon URL
		existing.IconURL = "/uploads/skills/" + fileName
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
		if file != nil {
			os.Remove(filepath.Join(s.uploadPath, filepath.Base(existing.IconURL)))
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

	// Hapus file icon jika ada
	if skill.IconURL != "" {
		fileName := filepath.Base(skill.IconURL)
		filePath := filepath.Join(s.uploadPath, fileName)
		os.Remove(filePath) // Ignore error jika file tidak ada
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
	repo       repo.CertificateRepository
	uploadPath string
}

func NewCertificateService(repo repo.CertificateRepository, uploadPath string) CertificateService {
	// Buat folder upload jika belum ada
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("Warning: gagal membuat folder upload certificate: %v\n", err)
	}
	return &certificateService{
		repo:       repo,
		uploadPath: uploadPath,
	}
}

func (s *certificateService) validateFile(file *multipart.FileHeader) error {
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
		".pdf":  true, // Untuk certificate mungkin butuh PDF juga
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		return errors.New("tipe file tidak diizinkan. File yang diizinkan: jpg, jpeg, png, webp, pdf")
	}

	return nil
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
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("certificate_%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.uploadPath, fileName)

	// Simpan file
	if err := ctx.SaveUploadedFile(file, filePath); err != nil {
		return nil, fmt.Errorf("gagal menyimpan file: %v", err)
	}

	// Parse issue date
	var issueDate time.Time
	if form.IssueDate != "" {
		parsedDate, err := time.Parse("2006-01-02", form.IssueDate)
		if err != nil {
			// Cleanup file jika parsing gagal
			os.Remove(filePath)
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
		ImageURL:      "/uploads/certificates/" + fileName, // Relative path
		IssueDate:     issueDate,
		Issuer:        form.Issuer,
		CredentialURL: form.CredentialURL,
		DisplayOrder:  form.DisplayOrder,
	}

	// Save to database
	if err := s.repo.Create(cert); err != nil {
		// Cleanup file jika gagal save ke database
		os.Remove(filePath)
		return nil, fmt.Errorf("gagal menyimpan data sertifikat: %v", err)
	}

	return s.convertCertToResponse(cert), nil
}

// Method lainnya tetap sama...
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
		return errors.New("ID sertif tidak valid")
	}

	// Check if project exists dan ambil datanya
	existingsertif, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("sertif tidak ditemukan")
	}

	// Delete dari database terlebih dahulu
	err = s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("gagal menghapus sertif: %v", err)
	}

	// Hapus file image jika ada dan bukan default
	if existingsertif.ImageURL != "" && existingsertif.ImageURL != "#" {
		fileName := filepath.Base(existingsertif.ImageURL)
		filePath := filepath.Join(s.uploadPath, fileName)

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
// EDUCATION SERVICE
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
// TESTIMONIALS SERVICE
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
// BLOG SERVICE
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
// HELPER FUNCTIONS
// ============================

func convertSkillToResponse(skill *model.Skill) *model.SkillResponse {
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

func convertCertToResponse(cert *model.Certificate) *model.CertificateResponse {
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

// ============================
// SECTIONS SERVICE
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
// SOCIAL LINKS SERVICE
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
// SETTINGS SERVICE
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
