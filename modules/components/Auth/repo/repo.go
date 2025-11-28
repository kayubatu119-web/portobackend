package userepo

import (
	"database/sql"
	"errors"
	modelsuser "gintugas/modules/components/Auth/model"

	"github.com/google/uuid"
)

type Repository interface {
	// CreateKategoriRepository(kategori Kategori) (result []Kategori, err error)
	GetAllUsersRepository() (result []modelsuser.User, err error)
	GetUsersRepository(id uuid.UUID) (modelsuser.User, error)
	GetUserByIDRepository(id uuid.UUID) (modelsuser.User, error)
	DeleteUsersRepository(id uuid.UUID) (err error)
	UpdateUsersRepository(users modelsuser.User) (modelsuser.User, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// func (r *repository) CreateKategoriRepository(kategori Kategori) (result []Kategori, err error) {
// 	query := `
//         INSERT INTO kategori (nama_kategori, description_kategori)
//         VALUES ($1, $2)
//         RETURNING id_kategori
//     `

// 	var id int
// 	err = r.db.QueryRow(query, kategori.NAMA_KATEGORI, kategori.DESCRIPTION_KATEGORI).Scan(&id)
// 	if err != nil {
// 		return nil, err
// 	}

// 	kategori.ID_KATEGORI = id

// 	allkategori, err := r.GetAllKategoriRepository()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return allkategori, nil
// }

func (r *repository) GetAllUsersRepository() (result []modelsuser.User, err error) {
	query := "SELECT id, username, email, role FROM users ORDER BY id"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []modelsuser.User
	for rows.Next() {
		var user modelsuser.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *repository) GetUsersRepository(id uuid.UUID) (modelsuser.User, error) {
	var user modelsuser.User

	query := "SELECT id, username, email, role, created_at FROM users WHERE id = $1"
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return modelsuser.User{}, errors.New("users tidak ditemukan")
		}
		return modelsuser.User{}, err
	}

	return user, nil
}

func (r *repository) GetUserByIDRepository(id uuid.UUID) (modelsuser.User, error) {
	sql := "SELECT id, username, email, password_hash, role FROM users WHERE id = $1"

	var user modelsuser.User
	err := r.db.QueryRow(sql, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role)

	if err != nil {
		return modelsuser.User{}, errors.New("user tidak ditemukan: " + err.Error())
	}

	return user, nil
}

func (r *repository) UpdateUsersRepository(users modelsuser.User) (modelsuser.User, error) {
	sql := "UPDATE users SET username = $1, email = $2, password_hash = $3, role = $4, updated_at = $5  WHERE id = $6 RETURNING id, username, email, password_hash, role, updated_at"

	var updateusers modelsuser.User
	err := r.db.QueryRow(sql, users.Username, users.Email, users.Password, users.Role, users.UpdatedAt, users.ID).
		Scan(&updateusers.ID, &updateusers.Username, &updateusers.Email, &updateusers.Password, &updateusers.Role, &updateusers.UpdatedAt)

	if err != nil {
		return modelsuser.User{}, errors.New("gagal mengupdate users: " + err.Error())
	}

	return updateusers, nil
}

func (r *repository) DeleteUsersRepository(id uuid.UUID) (err error) {

	sql := "DELETE FROM users WHERE id = $1"

	result, err := r.db.Exec(sql, id)
	if err != nil {
		return errors.New("gagal menghapus users: " + err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("gagal memeriksa rows affected: " + err.Error())
	}

	if rowsAffected == 0 {
		return errors.New("users tidak ditemukan")
	}

	return nil
}
