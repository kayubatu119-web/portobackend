package projectrepo

import (
	"database/sql"
	"errors"
	. "gintugas/modules/components/Project/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateProjekRepository(projek Project) (Project, error)
	GetAllProjekRepository() ([]Project, error)
	GetProjekRepository(id uuid.UUID) (Project, error)
	UpdateProjekRepository(projek Project) (Project, error)
	DeleteProjekRepository(id uuid.UUID) error
	GetProjekWithTagsRepository(id uuid.UUID) (Project, error)
	GetAllProjekWithTagsRepository() ([]Project, error)
	GetAllTagsRepository() (result []ProjectTag, err error)
}

type TagsRepository interface {
	CreateTags(Tags *ProjectTag) error
}

type repository struct {
	db *sql.DB
}

type tagsRepository struct {
	db *gorm.DB
}

func NewTagsRepository(db *gorm.DB) TagsRepository {
	return &tagsRepository{
		db: db,
	}
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAllTagsRepository() (result []ProjectTag, err error) {
	query := "SELECT id, name, color FROM project_tags ORDER BY id"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []ProjectTag
	for rows.Next() {
		var tag ProjectTag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *repository) CreateProjekRepository(projek Project) (Project, error) {
	query := `
		INSERT INTO portfolio_projects 
		(title, description, image_url, demo_url, code_url, display_order, is_featured, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		projek.Title,
		projek.Description,
		projek.ImageURL,
		projek.DemoURL,
		projek.CodeURL,
		projek.DisplayOrder,
		projek.IsFeatured,
		projek.Status,
	).Scan(&projek.ID, &projek.CreatedAt, &projek.UpdatedAt)

	if err != nil {
		return Project{}, err
	}

	return projek, nil
}

func (r *repository) GetAllProjekRepository() ([]Project, error) {
	query := `
		SELECT id, title, description, image_url, demo_url, code_url, 
		       display_order, is_featured, status, created_at, updated_at
		FROM portfolio_projects 
		ORDER BY display_order ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		err := rows.Scan(
			&project.ID,
			&project.Title,
			&project.Description,
			&project.ImageURL,
			&project.DemoURL,
			&project.CodeURL,
			&project.DisplayOrder,
			&project.IsFeatured,
			&project.Status,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (r *repository) GetProjekRepository(id uuid.UUID) (Project, error) {
	query := `
		SELECT id, title, description, image_url, demo_url, code_url, 
		       display_order, is_featured, status, created_at, updated_at
		FROM portfolio_projects 
		WHERE id = $1
	`

	var project Project
	err := r.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&project.ImageURL,
		&project.DemoURL,
		&project.CodeURL,
		&project.DisplayOrder,
		&project.IsFeatured,
		&project.Status,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Project{}, errors.New("project not found")
		}
		return Project{}, err
	}

	return project, nil
}

func (r *repository) UpdateProjekRepository(projek Project) (Project, error) {
	query := `
		UPDATE portfolio_projects 
		SET title = $1, description = $2, image_url = $3, demo_url = $4, 
		    code_url = $5, display_order = $6, is_featured = $7, status = $8,
			updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		projek.Title,
		projek.Description,
		projek.ImageURL,
		projek.DemoURL,
		projek.CodeURL,
		projek.DisplayOrder,
		projek.IsFeatured,
		projek.Status,
		projek.ID,
	).Scan(&projek.UpdatedAt)

	if err != nil {
		return Project{}, err
	}

	return projek, nil
}

func (r *repository) DeleteProjekRepository(id uuid.UUID) error {
	query := `DELETE FROM portfolio_projects WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("project not found")
	}

	return nil
}

func (r *repository) GetProjekWithTagsRepository(id uuid.UUID) (Project, error) {
	// First get project
	project, err := r.GetProjekRepository(id)
	if err != nil {
		return Project{}, err
	}

	// Then get tags for this project
	tagsQuery := `
		SELECT pt.id, pt.name, pt.color
		FROM project_tags pt
		INNER JOIN project_tag_relations ptr ON pt.id = ptr.tag_id
		WHERE ptr.project_id = $1
	`

	tagRows, err := r.db.Query(tagsQuery, id)
	if err != nil {
		return project, err
	}
	defer tagRows.Close()

	var tags []ProjectTag
	for tagRows.Next() {
		var tag ProjectTag
		err := tagRows.Scan(&tag.ID, &tag.Name, &tag.Color)
		if err != nil {
			return project, err
		}
		tags = append(tags, tag)
	}

	project.Tags = tags
	return project, nil
}

func (r *repository) GetAllProjekWithTagsRepository() ([]Project, error) {
	// Query untuk mendapatkan semua projects
	projectQuery := `
		SELECT id, title, description, image_url, demo_url, code_url, 
		       display_order, is_featured, status, created_at, updated_at
		FROM portfolio_projects 
		ORDER BY display_order ASC
	`

	projectRows, err := r.db.Query(projectQuery)
	if err != nil {
		return nil, err
	}
	defer projectRows.Close()

	var projects []Project
	projectMap := make(map[uuid.UUID]*Project) // Untuk mapping cepat

	// Scan semua projects
	for projectRows.Next() {
		var project Project
		err := projectRows.Scan(
			&project.ID,
			&project.Title,
			&project.Description,
			&project.ImageURL,
			&project.DemoURL,
			&project.CodeURL,
			&project.DisplayOrder,
			&project.IsFeatured,
			&project.Status,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Initialize tags slice
		project.Tags = []ProjectTag{}
		projects = append(projects, project)
		projectMap[project.ID] = &projects[len(projects)-1]
	}

	if len(projects) == 0 {
		return projects, nil
	}

	// Query untuk mendapatkan semua tags dari semua projects sekaligus (lebih efisien)
	tagsQuery := `
		SELECT 
			ptr.project_id,
			pt.id,
			pt.name,
			pt.color,
			pt.created_at
		FROM project_tag_relations ptr
		INNER JOIN project_tags pt ON ptr.tag_id = pt.id
		INNER JOIN portfolio_projects pp ON ptr.project_id = pp.id
		ORDER BY ptr.project_id, pt.name
	`

	tagRows, err := r.db.Query(tagsQuery)
	if err != nil {
		return projects, err // Return projects tanpa tags jika error
	}
	defer tagRows.Close()

	// Map tags ke projects yang sesuai
	for tagRows.Next() {
		var projectID uuid.UUID
		var tag ProjectTag

		err := tagRows.Scan(
			&projectID,
			&tag.ID,
			&tag.Name,
			&tag.Color,
			&tag.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Tambahkan tag ke project yang sesuai
		if project, exists := projectMap[projectID]; exists {
			project.Tags = append(project.Tags, tag)
		}
	}

	return projects, nil
}

func (r *tagsRepository) CreateTags(Tags *ProjectTag) error {
	return r.db.Create(Tags).Error
}
