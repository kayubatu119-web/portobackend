package projectservice

import (
	"fmt"
	"gintugas/modules/utils"
	"mime/multipart"
	"path/filepath"
	"strings"
)

// UploadServiceWrapper adalah interface untuk abstraksi upload service
type UploadServiceWrapper interface {
	UploadFile(file *multipart.FileHeader, folder string) (string, error)
	DeleteFile(fileURL string) error
	ValidateFile(file *multipart.FileHeader, maxSizeMB int64, allowedExts []string) error
}

// SupabaseUploadWrapper adalah wrapper untuk Supabase Upload Service
type SupabaseUploadWrapper struct {
	service *utils.SupabaseUploadService
}

func NewSupabaseUploadWrapper(service *utils.SupabaseUploadService) *SupabaseUploadWrapper {
	return &SupabaseUploadWrapper{
		service: service,
	}
}

func (s *SupabaseUploadWrapper) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	if err := s.ValidateFile(file, 10, []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg", ".pdf"}); err != nil {
		return "", err
	}

	return s.service.UploadFile(file, folder)
}

func (s *SupabaseUploadWrapper) DeleteFile(fileURL string) error {
	return s.service.DeleteFile(fileURL)
}

func (s *SupabaseUploadWrapper) ValidateFile(file *multipart.FileHeader, maxSizeMB int64, allowedExts []string) error {
	if file == nil {
		return nil
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

// LocalUploadWrapper adalah wrapper untuk Local Upload Service
type LocalUploadWrapper struct {
	service *utils.LocalUploadService
}

func NewLocalUploadWrapper(service *utils.LocalUploadService) *LocalUploadWrapper {
	return &LocalUploadWrapper{
		service: service,
	}
}

func (s *LocalUploadWrapper) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	return s.service.UploadFile(file, folder)
}

func (s *LocalUploadWrapper) DeleteFile(fileURL string) error {
	return s.service.DeleteFile(fileURL)
}

func (s *LocalUploadWrapper) ValidateFile(file *multipart.FileHeader, maxSizeMB int64, allowedExts []string) error {
	if file == nil {
		return nil
	}
	return nil
}
