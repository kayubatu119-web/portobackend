package serviceroute

import (
	"database/sql"
	"fmt"
	userrepo "gintugas/modules/components/Auth/repo"
	userservice "gintugas/modules/components/Auth/service-user"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAllUsersRouter godoc
// @Summary Get semua users
// @Description Mendapatkan daftar semua users (hanya admin)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/admin/users [get]
func GetAllUsersRouter(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			usersrepo = userrepo.NewRepository(db)
			usersSrv  = userservice.NewService(usersrepo)
		)

		users, err := usersSrv.GetAllUsersService(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("successfully get all users data"),
			"Users":   users,
		})
	}
}

// GetUsersRouter godoc
// @Summary Get user by ID
// @Description Mendapatkan detail user berdasarkan ID (hanya admin)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/admin/users/{id} [get]
func GetUsersRouter(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			usersrepo = userrepo.NewRepository(db)
			usersSrv  = userservice.NewService(usersrepo)
		)

		users, err := usersSrv.GetUserService(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("successfully get users data"),
			"Users":   users,
		})
	}
}

// UpdateUsersRouter godoc
// @Summary Update user
// @Description Update data user (hanya admin)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param input body map[string]interface{} true "Data user"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/admin/users/{id} [put]
func UpdateUsersRouter(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			usersrepo = userrepo.NewRepository(db)
			usersSrv  = userservice.NewService(usersrepo)
		)

		users, err := usersSrv.UpdateUserService(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Data users Berhasil di Update"),
			"Users":   users,
		})
	}
}

// DeleteUsersRouter godoc
// @Summary Delete user
// @Description Hapus user (hanya admin)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/admin/users/{id} [delete]
func DeleteUsersRouter(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			usersrepo = userrepo.NewRepository(db)
			usersSrv  = userservice.NewService(usersrepo)
		)

		err := usersSrv.DeleteUserService(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "users berhasil dihapus",
		})
	}
}
