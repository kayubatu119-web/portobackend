package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	storage "github.com/supabase-community/storage-go"
)

type SupabaseUploadService struct {
	client      *storage.Client
	bucket      string
	supabaseURL string // Untuk generate public URL
	projectID   string // Extract dari supabaseURL
}

// NewSupabaseUploadService membuat instance baru
// supabaseURL format: https://[PROJECT_ID].supabase.co
// Contoh: https://yiujndqqbacipqozosdm.supabase.co
func NewSupabaseUploadService(supabaseURL, supabaseKey, bucket string) *SupabaseUploadService {
	client := storage.NewClient(supabaseURL, supabaseKey, nil)

	// Extract project ID dari supabaseURL
	// https://yiujndqqbacipqozosdm.supabase.co -> yiujndqqbacipqozosdm
	projectID := extractProjectID(supabaseURL)

	return &SupabaseUploadService{
		client:      client,
		bucket:      bucket,
		supabaseURL: supabaseURL,
		projectID:   projectID,
	}
}

func extractProjectID(supabaseURL string) string {
	// Hapus protocol
	url := strings.TrimPrefix(supabaseURL, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Ambil bagian sebelum .supabase.co
	parts := strings.Split(url, ".supabase.co")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// ValidateFile memvalidasi file sebelum upload
func (s *SupabaseUploadService) ValidateFile(file *multipart.FileHeader, maxSizeMB int64, allowedExts []string) error {
	if file == nil {
		return errors.New("file tidak ditemukan")
	}

	// Check ukuran file
	maxSize := maxSizeMB * 1024 * 1024
	if file.Size > maxSize {
		return fmt.Errorf("ukuran file maksimal %dMB", maxSizeMB)
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		extStr := strings.Join(allowedExts, ", ")
		return fmt.Errorf("tipe file tidak diizinkan. File yang diizinkan: %s", extStr)
	}

	return nil
}

// UploadFile mengupload file ke Supabase Storage
// Mengembalikan public URL yang bisa diakses langsung
func (s *SupabaseUploadService) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	// Validasi file tidak kosong
	if file == nil {
		return "", errors.New("file tidak ditemukan")
	}

	// Buka file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %v", err)
	}
	defer src.Close()

	// Read file content
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file: %v", err)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Path di Supabase Storage
	storagePath := filename
	if folder != "" && folder != "/" {
		storagePath = fmt.Sprintf("%s/%s", strings.Trim(folder, "/"), filename)
	}

	// Upload ke Supabase Storage menggunakan bytes reader
	_, err = s.client.UploadFile(s.bucket, storagePath, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("gagal upload ke Supabase: %v", err)
	}

	// Generate public URL
	publicURL := s.GetPublicURL(storagePath)

	fmt.Printf("✅ File uploaded successfully: %s\n", publicURL)
	return publicURL, nil
}

// DeleteFile menghapus file dari Supabase Storage
func (s *SupabaseUploadService) DeleteFile(fileURL string) error {
	// Extract path dari URL
	// URL format: https://yiujndqqbacipqozosdm.supabase.co/storage/v1/object/public/bucket/path/to/file.jpg

	filePath := s.extractFilePathFromURL(fileURL)
	if filePath == "" {
		return fmt.Errorf("invalid file URL: %s", fileURL)
	}

	_, err := s.client.RemoveFile(s.bucket, []string{filePath})
	if err != nil {
		return fmt.Errorf("gagal menghapus file dari Supabase: %v", err)
	}

	fmt.Printf("✅ File deleted successfully: %s\n", filePath)
	return nil
}

// extractFilePathFromURL mengekstrak path dari public URL
func (s *SupabaseUploadService) extractFilePathFromURL(fileURL string) string {
	// Format: https://projectid.supabase.co/storage/v1/object/public/bucket/path/to/file

	// Cari "/storage/v1/object/public/"
	prefix := "/storage/v1/object/public/"
	idx := strings.Index(fileURL, prefix)
	if idx == -1 {
		return ""
	}

	// Ambil setelah prefix
	pathWithBucket := fileURL[idx+len(prefix):]

	// Buang bucket name di awal
	parts := strings.SplitN(pathWithBucket, "/", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}

// GetPublicURL menggenerate/construct URL public untuk file
func (s *SupabaseUploadService) GetPublicURL(filePath string) string {
	if s.projectID == "" {
		return "" // Return empty jika project ID tidak bisa di-extract
	}

	// Remove leading slashes
	filePath = strings.TrimPrefix(filePath, "/")

	return fmt.Sprintf("https://%s.supabase.co/storage/v1/object/public/%s/%s",
		s.projectID,
		s.bucket,
		filePath,
	)
}

// UploadBytes upload dari bytes array
// Berguna untuk image compression, base64, dll
func (s *SupabaseUploadService) UploadBytes(data []byte, filename, folder string) (string, error) {
	if len(data) == 0 {
		return "", errors.New("data kosong")
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg" // default extension
	}
	uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	storagePath := uniqueName
	if folder != "" && folder != "/" {
		storagePath = fmt.Sprintf("%s/%s", strings.Trim(folder, "/"), uniqueName)
	}

	// Upload ke Supabase
	_, err := s.client.UploadFile(s.bucket, storagePath, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("gagal upload bytes ke Supabase: %v", err)
	}

	publicURL := s.GetPublicURL(storagePath)
	return publicURL, nil
}

// FileExists mengecek apakah file sudah ada di storage
func (s *SupabaseUploadService) FileExists(filePath string) (bool, error) {
	// Gunakan list files untuk check existence
	// Ini adalah workaround karena storage-go SDK terbatas
	files, err := s.client.ListFiles(s.bucket, "", storage.FileSearchOptions{})
	if err != nil {
		return false, err
	}

	for _, file := range files {
		if file.Name == filePath {
			return true, nil
		}
	}

	return false, nil
}

// ListFiles menampilkan semua files di folder tertentu
func (s *SupabaseUploadService) ListFiles(folder string) ([]map[string]interface{}, error) {
	files, err := s.client.ListFiles(s.bucket, folder, storage.FileSearchOptions{})
	if err != nil {
		return nil, fmt.Errorf("gagal list files: %v", err)
	}

	var result []map[string]interface{}

	folder = strings.Trim(folder, "/")

	for _, file := range files {
		// Filter by folder jika diperlukan
		if folder != "" {
			if !strings.HasPrefix(file.Name, folder+"/") {
				continue
			}
		}

		result = append(result, map[string]interface{}{
			"name":       file.Name,
			"id":         file.Id,
			"updated_at": file.UpdatedAt,
			"created_at": file.CreatedAt,
			"url":        s.GetPublicURL(file.Name),
		})
	}

	return result, nil
}

// ===== BACKWARD COMPATIBILITY =====
// Fungsi-fungsi ini untuk kompatibilitas dengan kode lama

type UploadService = SupabaseUploadService

func NewUploadService(supabaseURL, supabaseKey, bucket string) *SupabaseUploadService {
	return NewSupabaseUploadService(supabaseURL, supabaseKey, bucket)
}

// ===== LOCAL FILE SYSTEM UPLOAD =====
// Untuk fallback jika Supabase tidak available

type LocalUploadService struct {
	uploadPath string
}

func NewLocalUploadService(uploadPath string) *LocalUploadService {
	// Buat folder jika belum ada
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("⚠️ Warning: gagal membuat folder upload: %v\n", err)
	}
	return &LocalUploadService{
		uploadPath: uploadPath,
	}
}

// UploadFile simpan file ke local folder
func (s *LocalUploadService) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	if file == nil {
		return "", errors.New("file tidak ditemukan")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Path folder
	uploadDir := s.uploadPath
	if folder != "" {
		uploadDir = filepath.Join(s.uploadPath, folder)
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return "", fmt.Errorf("gagal membuat folder: %v", err)
		}
	}

	filePath := filepath.Join(uploadDir, filename)

	// Simpan file
	if err := os.Mkdir(uploadDir, 0755); err != nil && !os.IsExist(err) {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// Return relative path untuk serve via static
	if folder != "" {
		return fmt.Sprintf("/uploads/%s/%s", folder, filename), nil
	}
	return fmt.Sprintf("/uploads/%s", filename), nil
}

// DeleteFile hapus file dari local storage
func (s *LocalUploadService) DeleteFile(filePath string) error {
	// filePath format: /uploads/folder/filename
	// Convert to actual file path
	actualPath := filepath.Join(s.uploadPath, strings.TrimPrefix(filePath, "/uploads/"))

	if err := os.Remove(actualPath); err != nil {
		return fmt.Errorf("gagal menghapus file: %v", err)
	}

	return nil
}
