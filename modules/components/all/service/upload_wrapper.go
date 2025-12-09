package service

import (
	"gintugas/modules/utils"
	"mime/multipart"
	"strings"
)

// UploadServiceWrapper adalah interface untuk abstraksi upload service
// Bisa diimplementasikan dengan Supabase atau Local storage
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
	return s.service.UploadFile(file, folder)
}

func (s *SupabaseUploadWrapper) DeleteFile(fileURL string) error {
	return s.service.DeleteFile(fileURL)
}

func (s *SupabaseUploadWrapper) ValidateFile(file *multipart.FileHeader, maxSizeMB int64, allowedExts []string) error {
	return s.service.ValidateFile(file, maxSizeMB, allowedExts)
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
		return nil // Local service tidak strict validation
	}
	return nil
}

// Helper function untuk extract filename dari URL
func ExtractFilenameFromURL(fileURL string) string {
	parts := strings.Split(fileURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// Helper function untuk extract folder path dari URL
func ExtractFolderFromURL(fileURL string) string {
	// Untuk Supabase: https://project.supabase.co/storage/v1/object/public/portfolio/skills/uuid.jpg
	// Extract: skills

	if strings.Contains(fileURL, "/object/public/") {
		parts := strings.Split(fileURL, "/object/public/")
		if len(parts) > 1 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) > 1 {
				return pathParts[1] // Return folder name
			}
		}
	}
	return ""
}
