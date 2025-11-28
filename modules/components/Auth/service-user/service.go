package userservice

import (
	"errors"
	models "gintugas/modules/components/Auth/model"
	. "gintugas/modules/components/Auth/repo"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	// CreateKategoriService(ctx *gin.Context) (Kategori, error)
	GetAllUsersService(ctx *gin.Context) (result []models.User, err error)
	GetUserService(ctx *gin.Context) (result models.User, err error)
	UpdateUserService(ctx *gin.Context) (u models.User, err error)
	DeleteUserService(ctx *gin.Context) (err error)
}
type userService struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &userService{
		repository,
	}
}

// func (s *userService) CreateKategoriService(ctx *gin.Context) (Kategori, error) {
// 	var kategori Kategori

// 	if err := ctx.ShouldBindJSON(&kategori); err != nil {
// 		return kategori, err
// 	}

// 	if kategori.NAMA_KATEGORI == "" {
// 		return kategori, errors.New("nama kategori harus diisi")
// 	}

// 	if kategori.DESCRIPTION_KATEGORI == "" {
// 		return kategori, errors.New("deskripsi kategori harus diisi")
// 	}

// 	result, err := s.repository.CreateKategoriRepository(kategori)
// 	if err != nil {
// 		return kategori, errors.New("gagal menambahkan kategori: " + err.Error())
// 	}

// 	return result[kategori.ID_KATEGORI], nil
// }

func (s *userService) GetAllUsersService(ctx *gin.Context) (result []models.User, err error) {
	Users, err := s.repository.GetAllUsersRepository()
	if err != nil {
		return nil, errors.New("gagal mengambil data Users: " + err.Error())
	}

	return Users, nil
}

func (s *userService) GetUserService(ctx *gin.Context) (result models.User, err error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return models.User{}, errors.New("ID users tidak valid")
	}

	result, err = s.repository.GetUsersRepository(id)
	if err != nil {
		return models.User{}, errors.New("users tidak ditemukan")
	}

	return result, nil
}

func (s *userService) UpdateUserService(ctx *gin.Context) (u models.User, err error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return models.User{}, errors.New("ID users tidak valid")
	}

	existingUser, err := s.repository.GetUserByIDRepository(id)
	if err != nil {
		return models.User{}, err
	}

	var users models.User
	if err := ctx.ShouldBindJSON(&users); err != nil {
		return models.User{}, errors.New("data request tidak valid")
	}

	if strings.TrimSpace(users.Username) == "" {
		users.Username = existingUser.Username
	}

	if strings.TrimSpace(users.Email) == "" {
		users.Email = existingUser.Email
	}

	if strings.TrimSpace(users.Role) == "" {
		users.Role = existingUser.Role
	}

	if strings.TrimSpace(users.Password) != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(users.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.User{}, errors.New("gagal encrypt password")
		}
		users.Password = string(hashedPassword)
	} else {
		users.Password = existingUser.Password
	}

	users.ID = id

	u, err = s.repository.UpdateUsersRepository(users)
	if err != nil {
		return models.User{}, err
	}

	return u, nil
}

func (s *userService) DeleteUserService(ctx *gin.Context) (err error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errors.New("ID users tidak valid")
	}

	err = s.repository.DeleteUsersRepository(id)
	if err != nil {
		return err
	}

	return
}
