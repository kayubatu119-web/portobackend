package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SupabaseUploadService struct {
	supabaseURL string
	apiKey      string
	bucket      string
	client      *http.Client
}

func NewSupabaseUploadService(supabaseURL, apiKey, bucket string) *SupabaseUploadService {
	fmt.Println("🔧 Initializing Supabase Storage Service")
	fmt.Printf("   URL: %s\n", supabaseURL)
	fmt.Printf("   Bucket: %s\n", bucket)
	fmt.Printf("   Key available: %v\n", apiKey != "")

	// Test koneksi sederhana
	testURL := fmt.Sprintf("%s/storage/v1/bucket", strings.TrimSuffix(supabaseURL, "/"))
	req, _ := http.NewRequest("GET", testURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("⚠️  Initial test failed: %v\n", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			fmt.Println("✅ Connection test successful")
		} else {
			fmt.Printf("⚠️  Connection test returned status: %d\n", resp.StatusCode)
		}
	}

	return &SupabaseUploadService{
		supabaseURL: strings.TrimSuffix(supabaseURL, "/"),
		apiKey:      apiKey,
		bucket:      bucket,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// UploadFile mengupload file ke Supabase Storage menggunakan HTTP API langsung
func (s *SupabaseUploadService) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	if file == nil {
		return "", errors.New("file tidak ditemukan")
	}

	fmt.Printf("📤 Starting Supabase upload:\n")
	fmt.Printf("   File: %s (%.2f MB)\n", file.Filename, float64(file.Size)/(1024*1024))
	fmt.Printf("   Folder: %s\n", folder)

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
	if ext == "" {
		// Default extension berdasarkan content type
		contentType := file.Header.Get("Content-Type")
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		case "image/svg+xml":
			ext = ".svg"
		default:
			ext = ".dat"
		}
	}

	// Buat nama file unik
	uniqueID := uuid.New().String()
	filename := fmt.Sprintf("%s%s", uniqueID, ext)

	// Path di storage
	storagePath := filename
	if folder != "" {
		folder = strings.Trim(folder, "/")
		storagePath = fmt.Sprintf("%s/%s", folder, filename)
	}

	// Upload menggunakan HTTP API
	publicURL, err := s.uploadViaHTTP(fileBytes, storagePath, file)
	if err != nil {
		return "", fmt.Errorf("upload failed: %v", err)
	}

	fmt.Printf("✅ Upload successful!\n")
	fmt.Printf("   Public URL: %s\n", publicURL)

	return publicURL, nil
}

// uploadViaHTTP menggunakan HTTP API langsung ke Supabase
func (s *SupabaseUploadService) uploadViaHTTP(data []byte, storagePath string, file *multipart.FileHeader) (string, error) {
	// URL format: https://<project>.supabase.co/storage/v1/object/<bucket>/<path>
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s",
		s.supabaseURL,
		s.bucket,
		storagePath,
	)

	fmt.Printf("   Upload URL: %s\n", uploadURL)
	fmt.Printf("   File size: %d bytes\n", len(data))

	// Create request
	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("gagal membuat request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	// Tentukan content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		// Coba deteksi dari extension
		ext := filepath.Ext(file.Filename)
		switch strings.ToLower(ext) {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".webp":
			contentType = "image/webp"
		case ".svg":
			contentType = "image/svg+xml"
		case ".pdf":
			contentType = "application/pdf"
		default:
			contentType = "application/octet-stream"
		}
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Cache-Control", "public, max-age=31536000")

	// Untuk file gambar, tambahkan content-disposition
	if strings.HasPrefix(contentType, "image/") {
		req.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.Filename))
	}

	// Kirim request
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal mengirim request: %v", err)
	}
	defer resp.Body.Close()

	// Baca response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca response: %v", err)
	}

	fmt.Printf("   Response Status: %d\n", resp.StatusCode)

	// Cek response
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		// Success! Return public URL
		publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
			s.supabaseURL,
			s.bucket,
			storagePath,
		)
		return publicURL, nil
	}

	// Error handling
	errorMsg := string(respBody)
	fmt.Printf("   ❌ Error Response: %s\n", errorMsg)

	if resp.StatusCode == 401 {
		return "", fmt.Errorf("authentication failed - check your service role key")
	} else if resp.StatusCode == 403 {
		return "", fmt.Errorf("permission denied - check bucket permissions")
	} else if resp.StatusCode == 404 {
		return "", fmt.Errorf("bucket not found: %s", s.bucket)
	} else if resp.StatusCode == 413 {
		return "", fmt.Errorf("file too large")
	} else {
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, errorMsg)
	}
}

// DeleteFile menghapus file dari Supabase Storage
func (s *SupabaseUploadService) DeleteFile(fileURL string) error {
	// Ekstrak path dari URL
	path := s.extractFilePathFromURL(fileURL)
	if path == "" {
		return fmt.Errorf("invalid file URL format: %s", fileURL)
	}

	fmt.Printf("🗑️ Deleting file from Supabase: %s\n", path)

	// URL untuk delete
	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s",
		s.supabaseURL,
		s.bucket,
		path,
	)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("gagal membuat delete request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("gagal menghapus file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("delete failed with status %d", resp.StatusCode)
	}

	fmt.Printf("✅ File deleted successfully\n")
	return nil
}

// extractFilePathFromURL mengekstrak path file dari URL publik
func (s *SupabaseUploadService) extractFilePathFromURL(fileURL string) string {
	// Pattern 1: https://project.supabase.co/storage/v1/object/public/bucket/path/to/file
	publicPrefix := "/storage/v1/object/public/"

	idx := strings.Index(fileURL, publicPrefix)
	if idx != -1 {
		// Ambil bagian setelah prefix
		pathWithBucket := fileURL[idx+len(publicPrefix):]

		// Hilangkan bucket dari path
		if strings.HasPrefix(pathWithBucket, s.bucket+"/") {
			return strings.TrimPrefix(pathWithBucket, s.bucket+"/")
		}
	}

	// Pattern 2: coba ekstrak langsung dari path
	urlParts := strings.Split(fileURL, "/")
	for i, part := range urlParts {
		if part == s.bucket && i+1 < len(urlParts) {
			// Gabungkan bagian setelah bucket
			return strings.Join(urlParts[i+1:], "/")
		}
	}

	return ""
}

// GetPublicURL menghasilkan URL publik untuk file
func (s *SupabaseUploadService) GetPublicURL(filePath string) string {
	filePath = strings.TrimPrefix(filePath, "/")
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		s.supabaseURL,
		s.bucket,
		filePath,
	)
}

// UploadBytes untuk upload data byte langsung
func (s *SupabaseUploadService) UploadBytes(data []byte, filename, folder string) (string, error) {
	if len(data) == 0 {
		return "", errors.New("data kosong")
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".dat"
	}
	uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Build storage path
	storagePath := uniqueName
	if folder != "" {
		folder = strings.Trim(folder, "/")
		storagePath = fmt.Sprintf("%s/%s", folder, uniqueName)
	}

	// Upload
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s",
		s.supabaseURL,
		s.bucket,
		storagePath,
	)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	return s.GetPublicURL(storagePath), nil
}

// ============================================
// LOCAL UPLOAD SERVICE (untuk development)
// ============================================

type LocalUploadService struct {
	uploadPath string
}

func NewLocalUploadService(uploadPath string) *LocalUploadService {
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("⚠️ Warning: gagal membuat folder upload: %v\n", err)
	}
	fmt.Printf("📁 Local upload path: %s\n", uploadPath)
	return &LocalUploadService{uploadPath: uploadPath}
}

func (s *LocalUploadService) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	if file == nil {
		return "", errors.New("file tidak ditemukan")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".dat"
	}
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Tentukan direktori tujuan
	uploadDir := s.uploadPath
	if folder != "" {
		uploadDir = filepath.Join(s.uploadPath, folder)
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return "", fmt.Errorf("gagal membuat folder: %v", err)
		}
	}

	// Path lengkap file
	filePath := filepath.Join(uploadDir, filename)

	// Buka file source
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %v", err)
	}
	defer src.Close()

	// Buat file destination
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file: %v", err)
	}
	defer dst.Close()

	// Copy data
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("gagal menyalin file: %v", err)
	}

	// Return relative URL
	if folder != "" {
		return fmt.Sprintf("/uploads/%s/%s", folder, filename), nil
	}
	return fmt.Sprintf("/uploads/%s", filename), nil
}

func (s *LocalUploadService) DeleteFile(fileURL string) error {
	// Hapus prefix /uploads/
	relativePath := strings.TrimPrefix(fileURL, "/uploads/")
	if relativePath == fileURL {
		return fmt.Errorf("invalid file URL: %s", fileURL)
	}

	// Path absolut
	absolutePath := filepath.Join(s.uploadPath, relativePath)

	// Hapus file
	if err := os.Remove(absolutePath); err != nil {
		return fmt.Errorf("gagal menghapus file: %v", err)
	}

	return nil
}

// ============================================
// UPLOAD WRAPPERS
// ============================================

type UploadServiceWrapper interface {
	UploadFile(file *multipart.FileHeader, folder string) (string, error)
	DeleteFile(fileURL string) error
}

// SupabaseUploadWrapper
type SupabaseUploadWrapper struct {
	service *SupabaseUploadService
}

func NewSupabaseUploadWrapper(service *SupabaseUploadService) *SupabaseUploadWrapper {
	return &SupabaseUploadWrapper{service: service}
}

func (s *SupabaseUploadWrapper) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	return s.service.UploadFile(file, folder)
}

func (s *SupabaseUploadWrapper) DeleteFile(fileURL string) error {
	return s.service.DeleteFile(fileURL)
}

// LocalUploadWrapper
type LocalUploadWrapper struct {
	service *LocalUploadService
}

func NewLocalUploadWrapper(service *LocalUploadService) *LocalUploadWrapper {
	return &LocalUploadWrapper{service: service}
}

func (l *LocalUploadWrapper) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	return l.service.UploadFile(file, folder)
}

func (l *LocalUploadWrapper) DeleteFile(fileURL string) error {
	return l.service.DeleteFile(fileURL)
}
