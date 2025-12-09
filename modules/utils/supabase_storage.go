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

	"github.com/google/uuid"
	storage "github.com/supabase-community/storage-go"
)

type SupabaseUploadService struct {
	client      *storage.Client
	bucket      string
	supabaseURL string
	projectID   string
	apiKey      string // ⭐ TAMBAHKAN INI
}

func NewSupabaseUploadService(supabaseURL, supabaseKey, bucket string) *SupabaseUploadService {
	client := storage.NewClient(supabaseURL, supabaseKey, nil)
	projectID := extractProjectID(supabaseURL)

	fmt.Printf("🔧 Initializing Supabase Storage\n")
	fmt.Printf("   URL: %s\n", supabaseURL)
	fmt.Printf("   Project ID: %s\n", projectID)
	fmt.Printf("   Bucket: %s\n", bucket)
	fmt.Printf("   Key (first 10 chars): %s\n", supabaseKey[:10])

	// Test connection
	testPath := "test-connection.txt"
	testData := []byte("test connection")
	_, err := client.UploadFile(bucket, testPath, bytes.NewReader(testData))
	if err != nil {
		fmt.Printf("⚠️  Initial connection test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Connection test successful\n")
		// Cleanup test file
		client.RemoveFile(bucket, []string{testPath})
	}

	return &SupabaseUploadService{
		client:      client,
		bucket:      bucket,
		supabaseURL: supabaseURL,
		projectID:   projectID,
		apiKey:      supabaseKey, // ⭐ SIMPAN API KEY
	}
}

func extractProjectID(supabaseURL string) string {
	url := strings.TrimPrefix(supabaseURL, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")

	// Remove path and query params
	parts := strings.Split(url, "/")
	hostname := parts[0]

	// Extract project ID from *.supabase.co
	if strings.HasSuffix(hostname, ".supabase.co") {
		projectID := strings.TrimSuffix(hostname, ".supabase.co")
		return projectID
	}

	return hostname
}

// UploadFile mengupload file ke Supabase Storage
func (s *SupabaseUploadService) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
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
	if ext == "" {
		ext = ".png"
	}
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Path di Supabase Storage
	storagePath := filename
	if folder != "" {
		folder = strings.Trim(folder, "/")
		storagePath = fmt.Sprintf("%s/%s", folder, filename)
	}

	fmt.Printf("📤 Uploading file:\n")
	fmt.Printf("   Original: %s\n", file.Filename)
	fmt.Printf("   Storage Path: %s\n", storagePath)
	fmt.Printf("   Size: %d bytes\n", len(fileBytes))
	fmt.Printf("   Content-Type: %s\n", file.Header.Get("Content-Type"))

	// Upload ke Supabase Storage menggunakan HTTP API langsung
	// (Lebih reliable daripada library client)
	publicURL, err := s.uploadViaHTTP(fileBytes, storagePath, file.Header.Get("Content-Type"))
	if err != nil {
		return "", fmt.Errorf("upload failed: %v", err)
	}

	fmt.Printf("✅ Upload successful: %s\n", publicURL)
	return publicURL, nil
}

// uploadViaHTTP menggunakan HTTP API langsung
func (s *SupabaseUploadService) uploadViaHTTP(data []byte, path, contentType string) (string, error) {
	if contentType == "" {
		// Determine content type from extension
		ext := filepath.Ext(path)
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
		default:
			contentType = "application/octet-stream"
		}
	}

	// URL format: https://project-id.supabase.co/storage/v1/object/bucket/path
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s",
		strings.TrimSuffix(s.supabaseURL, "/"),
		s.bucket,
		path,
	)

	fmt.Printf("   Upload URL: %s\n", uploadURL)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// ⚠️ PERBAIKAN: Gunakan API key langsung, bukan s.client.AccessToken
	// Anda perlu menyimpan API key di struct
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		fmt.Printf("   ❌ Upload failed: %s\n", resp.Status)
		fmt.Printf("   Response: %s\n", string(body))
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Construct public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		strings.TrimSuffix(s.supabaseURL, "/"),
		s.bucket,
		path,
	)

	return publicURL, nil
}

func (s *SupabaseUploadService) DeleteFile(fileURL string) error {
	filePath := s.extractFilePathFromURL(fileURL)
	if filePath == "" {
		return fmt.Errorf("invalid file URL: %s", fileURL)
	}

	_, err := s.client.RemoveFile(s.bucket, []string{filePath})
	if err != nil {
		return fmt.Errorf("gagal menghapus file dari Supabase: %v", err)
	}

	fmt.Printf("🗑️ File deleted: %s\n", filePath)
	return nil
}

func (s *SupabaseUploadService) extractFilePathFromURL(fileURL string) string {
	// Pattern: https://project-id.supabase.co/storage/v1/object/public/bucket/path/to/file
	prefixes := []string{
		"/storage/v1/object/public/",
		"storage/v1/object/public/",
	}

	for _, prefix := range prefixes {
		idx := strings.Index(fileURL, prefix)
		if idx != -1 {
			pathWithBucket := fileURL[idx+len(prefix):]
			parts := strings.SplitN(pathWithBucket, "/", 2)
			if len(parts) == 2 && parts[0] == s.bucket {
				return parts[1]
			}
		}
	}

	return ""
}

func (s *SupabaseUploadService) GetPublicURL(filePath string) string {
	filePath = strings.TrimPrefix(filePath, "/")
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		strings.TrimSuffix(s.supabaseURL, "/"),
		s.bucket,
		filePath,
	)
}

func (s *SupabaseUploadService) UploadBytes(data []byte, filename, folder string) (string, error) {
	if len(data) == 0 {
		return "", errors.New("data kosong")
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	storagePath := uniqueName
	if folder != "" && folder != "/" {
		folder = strings.Trim(folder, "/")
		storagePath = fmt.Sprintf("%s/%s", folder, uniqueName)
	}

	_, err := s.client.UploadFile(s.bucket, storagePath, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	return s.GetPublicURL(storagePath), nil
}

// UploadServiceWrapper interface
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

// LocalUploadService
type LocalUploadService struct {
	uploadPath string
}

func NewLocalUploadService(uploadPath string) *LocalUploadService {
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("⚠️ Warning: gagal membuat folder upload: %v\n", err)
	}
	return &LocalUploadService{uploadPath: uploadPath}
}

func (s *LocalUploadService) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	if file == nil {
		return "", errors.New("file tidak ditemukan")
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	uploadDir := s.uploadPath
	if folder != "" {
		uploadDir = filepath.Join(s.uploadPath, folder)
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return "", err
		}
	}

	filePath := filepath.Join(uploadDir, filename)

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

	if folder != "" {
		return fmt.Sprintf("/uploads/%s/%s", folder, filename), nil
	}
	return fmt.Sprintf("/uploads/%s", filename), nil
}

func (s *LocalUploadService) DeleteFile(filePath string) error {
	actualPath := filepath.Join(s.uploadPath, strings.TrimPrefix(filePath, "/uploads/"))
	return os.Remove(actualPath)
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
