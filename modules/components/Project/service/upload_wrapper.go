package projectservice

import (
	"gintugas/modules/utils"
	"mime/multipart"
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
	return s.service.UploadFile(file, folder)
}

func (s *SupabaseUploadWrapper) DeleteFile(fileURL string) error {
	return s.service.DeleteFile(fileURL)
}

func (s *SupabaseUploadWrapper) ValidateFile(file *multipart.FileHeader, maxSizeMB int64, allowedExts []string) error {
	return s.ValidateFile(file, maxSizeMB, allowedExts)
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
