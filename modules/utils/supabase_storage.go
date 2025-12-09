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
}

func NewSupabaseUploadService(supabaseURL, supabaseKey, bucket string) *SupabaseUploadService {
	client := storage.NewClient(supabaseURL, supabaseKey, nil)
	projectID := extractProjectID(supabaseURL)

	fmt.Printf("🔧 Initializing Supabase Storage: %s, Bucket: %s\n", projectID, bucket)

	return &SupabaseUploadService{
		client:      client,
		bucket:      bucket,
		supabaseURL: supabaseURL,
		projectID:   projectID,
	}
}

func extractProjectID(supabaseURL string) string {
	url := strings.TrimPrefix(supabaseURL, "https://")
	url = strings.TrimPrefix(url, "http://")
	parts := strings.Split(url, ".supabase.co")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
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
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Path di Supabase Storage
	storagePath := filename
	if folder != "" && folder != "/" {
		folder = strings.Trim(folder, "/")
		storagePath = fmt.Sprintf("%s/%s", folder, filename)
	}

	fmt.Printf("📤 Uploading to: %s/%s (Size: %d bytes)\n", s.bucket, storagePath, len(fileBytes))

	// Upload ke Supabase Storage
	_, err = s.client.UploadFile(s.bucket, storagePath, bytes.NewReader(fileBytes))
	if err != nil {
		// Coba upload dengan HTTP langsung
		return s.uploadViaHTTP(fileBytes, storagePath, file.Header.Get("Content-Type"))
	}

	// Generate public URL
	publicURL := s.GetPublicURL(storagePath)
	fmt.Printf("✅ Upload successful: %s\n", publicURL)

	return publicURL, nil
}

// uploadViaHTTP alternatif upload method
func (s *SupabaseUploadService) uploadViaHTTP(data []byte, path, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.supabaseURL, s.bucket, path)

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s - %s", resp.Status, string(body))
	}

	return s.GetPublicURL(path), nil
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
	prefix := "/storage/v1/object/public/"
	idx := strings.Index(fileURL, prefix)
	if idx == -1 {
		return ""
	}

	pathWithBucket := fileURL[idx+len(prefix):]
	parts := strings.SplitN(pathWithBucket, "/", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}

func (s *SupabaseUploadService) GetPublicURL(filePath string) string {
	filePath = strings.TrimPrefix(filePath, "/")
	return fmt.Sprintf("https://%s.supabase.co/storage/v1/object/public/%s/%s",
		s.projectID,
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
