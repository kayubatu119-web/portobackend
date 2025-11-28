package repo

import (
	model "gintugas/modules/components/all/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================
// SKILLS REPOSITORY
// ============================

type SkillRepository interface {
	Create(skill *model.Skill) error
	GetByID(id uuid.UUID) (*model.Skill, error)
	Update(skill *model.Skill) error
	Delete(id uuid.UUID) error
	GetAll() ([]model.Skill, error)
	GetFeatured() ([]model.Skill, error)
	GetByCategory(category string) ([]model.Skill, error)
}

type skillRepository struct {
	db *gorm.DB
}

func NewSkillRepository(db *gorm.DB) SkillRepository {
	return &skillRepository{db: db}
}

func (r *skillRepository) Create(skill *model.Skill) error {
	return r.db.Create(skill).Error
}

func (r *skillRepository) GetByID(id uuid.UUID) (*model.Skill, error) {
	var skill model.Skill
	err := r.db.Where("id = ?", id).First(&skill).Error
	return &skill, err
}

func (r *skillRepository) Update(skill *model.Skill) error {
	return r.db.Save(skill).Error
}

func (r *skillRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Skill{}).Error
}

func (r *skillRepository) GetAll() ([]model.Skill, error) {
	var skills []model.Skill
	err := r.db.Order("display_order ASC, created_at DESC").Find(&skills).Error
	return skills, err
}

func (r *skillRepository) GetFeatured() ([]model.Skill, error) {
	var skills []model.Skill
	err := r.db.Where("is_featured = ?", true).Order("display_order ASC").Find(&skills).Error
	return skills, err
}

func (r *skillRepository) GetByCategory(category string) ([]model.Skill, error) {
	var skills []model.Skill
	err := r.db.Where("category = ?", category).Order("display_order ASC").Find(&skills).Error
	return skills, err
}

// ============================
// CERTIFICATES REPOSITORY
// ============================

type CertificateRepository interface {
	Create(cert *model.Certificate) error
	GetByID(id uuid.UUID) (*model.Certificate, error)
	Update(cert *model.Certificate) error
	Delete(id uuid.UUID) error
	GetAll() ([]model.Certificate, error)
}

type certificateRepository struct {
	db *gorm.DB
}

func NewCertificateRepository(db *gorm.DB) CertificateRepository {
	return &certificateRepository{db: db}
}

func (r *certificateRepository) Create(cert *model.Certificate) error {
	return r.db.Create(cert).Error
}

func (r *certificateRepository) GetByID(id uuid.UUID) (*model.Certificate, error) {
	var cert model.Certificate
	err := r.db.Where("id = ?", id).First(&cert).Error
	return &cert, err
}

func (r *certificateRepository) Update(cert *model.Certificate) error {
	return r.db.Save(cert).Error
}

func (r *certificateRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Certificate{}).Error
}

func (r *certificateRepository) GetAll() ([]model.Certificate, error) {
	var certs []model.Certificate
	err := r.db.Order("display_order ASC, created_at DESC").Find(&certs).Error
	return certs, err
}

// ============================
// EDUCATION REPOSITORY
// ============================

type EducationRepository interface {
	CreateWithAchievements(edu *model.Education) error
	GetByIDWithAchievements(id uuid.UUID) (*model.Education, error)
	UpdateWithAchievements(edu *model.Education) error
	DeleteWithAchievements(id uuid.UUID) error
	GetAllWithAchievements() ([]model.Education, error)
}

type educationRepository struct {
	db *gorm.DB
}

func NewEducationRepository(db *gorm.DB) EducationRepository {
	return &educationRepository{db: db}
}

func (r *educationRepository) CreateWithAchievements(edu *model.Education) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(edu).Error; err != nil {
			return err
		}

		for i := range edu.Achievements {
			edu.Achievements[i].EducationID = edu.ID
			edu.Achievements[i].ID = uuid.Nil
		}

		if len(edu.Achievements) > 0 {
			if err := tx.Create(&edu.Achievements).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *educationRepository) GetByIDWithAchievements(id uuid.UUID) (*model.Education, error) {
	var edu model.Education
	err := r.db.
		Preload("Achievements", func(db *gorm.DB) *gorm.DB {
			return db.Order("education_achievements.display_order ASC")
		}).
		Where("id = ?", id).
		First(&edu).Error
	return &edu, err
}

func (r *educationRepository) UpdateWithAchievements(edu *model.Education) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(edu).Error; err != nil {
			return err
		}

		if err := tx.Where("education_id = ?", edu.ID).Delete(&model.EducationAchievement{}).Error; err != nil {
			return err
		}

		if len(edu.Achievements) > 0 {
			for i := range edu.Achievements {
				edu.Achievements[i].EducationID = edu.ID
				edu.Achievements[i].ID = uuid.Nil
			}
			if err := tx.Create(&edu.Achievements).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *educationRepository) DeleteWithAchievements(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("education_id = ?", id).Delete(&model.EducationAchievement{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.Education{}).Error
	})
}

func (r *educationRepository) GetAllWithAchievements() ([]model.Education, error) {
	var educations []model.Education
	err := r.db.
		Preload("Achievements", func(db *gorm.DB) *gorm.DB {
			return db.Order("education_achievements.display_order ASC")
		}).
		Order("display_order ASC, created_at DESC").
		Find(&educations).Error
	return educations, err
}

// ============================
// TESTIMONIALS REPOSITORY
// ============================

type TestimonialRepository interface {
	Create(test *model.Testimonial) error
	GetByID(id uuid.UUID) (*model.Testimonial, error)
	Update(test *model.Testimonial) error
	Delete(id uuid.UUID) error
	GetAll() ([]model.Testimonial, error)
	GetFeatured() ([]model.Testimonial, error)
	GetByStatus(status string) ([]model.Testimonial, error)
}

type testimonialRepository struct {
	db *gorm.DB
}

func NewTestimonialRepository(db *gorm.DB) TestimonialRepository {
	return &testimonialRepository{db: db}
}

func (r *testimonialRepository) Create(test *model.Testimonial) error {
	return r.db.Create(test).Error
}

func (r *testimonialRepository) GetByID(id uuid.UUID) (*model.Testimonial, error) {
	var test model.Testimonial
	err := r.db.Where("id = ?", id).First(&test).Error
	return &test, err
}

func (r *testimonialRepository) Update(test *model.Testimonial) error {
	return r.db.Save(test).Error
}

func (r *testimonialRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Testimonial{}).Error
}

func (r *testimonialRepository) GetAll() ([]model.Testimonial, error) {
	var testimonials []model.Testimonial
	err := r.db.Order("display_order ASC, created_at DESC").Find(&testimonials).Error
	return testimonials, err
}

func (r *testimonialRepository) GetFeatured() ([]model.Testimonial, error) {
	var testimonials []model.Testimonial
	err := r.db.Where("is_featured = ?", true).Order("display_order ASC").Find(&testimonials).Error
	return testimonials, err
}

func (r *testimonialRepository) GetByStatus(status string) ([]model.Testimonial, error) {
	var testimonials []model.Testimonial
	err := r.db.Where("status = ?", status).Order("display_order ASC").Find(&testimonials).Error
	return testimonials, err
}

// ============================
// BLOG REPOSITORY
// ============================

type BlogRepository interface {
	CreateWithTags(post *model.BlogPost) error
	GetByIDWithTags(id uuid.UUID) (*model.BlogPost, error)
	GetBySlugWithTags(slug string) (*model.BlogPost, error)
	UpdateWithTags(post *model.BlogPost) error
	DeleteWithTags(id uuid.UUID) error
	GetAllWithTags() ([]model.BlogPost, error)
	GetPublishedWithTags() ([]model.BlogPost, error)
	IncrementViewCount(id uuid.UUID) error

	// Tag operations
	CreateTag(tag *model.BlogTag) error
	GetOrCreateTag(name string) (*model.BlogTag, error)
	GetAllTags() ([]model.BlogTag, error)
}

type blogRepository struct {
	db *gorm.DB
}

func NewBlogRepository(db *gorm.DB) BlogRepository {
	return &blogRepository{db: db}
}

func (r *blogRepository) CreateWithTags(post *model.BlogPost) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Handle tags first - get or create
		var processedTags []model.BlogTag
		if len(post.Tags) > 0 {
			for _, tag := range post.Tags {
				existingTag, err := r.getOrCreateTagTx(tx, tag.Name)
				if err != nil {
					return err
				}
				processedTags = append(processedTags, *existingTag)
			}
		}

		// Replace tags with processed ones
		post.Tags = processedTags

		// Create post with associations using GORM
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *blogRepository) GetByIDWithTags(id uuid.UUID) (*model.BlogPost, error) {
	var post model.BlogPost
	err := r.db.Preload("Tags").Where("id = ?", id).First(&post).Error
	return &post, err
}

func (r *blogRepository) GetBySlugWithTags(slug string) (*model.BlogPost, error) {
	var post model.BlogPost
	err := r.db.Preload("Tags").Where("slug = ?", slug).First(&post).Error
	return &post, err
}

func (r *blogRepository) UpdateWithTags(post *model.BlogPost) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Handle tags - get or create
		var processedTags []model.BlogTag
		if len(post.Tags) > 0 {
			for _, tag := range post.Tags {
				existingTag, err := r.getOrCreateTagTx(tx, tag.Name)
				if err != nil {
					return err
				}
				processedTags = append(processedTags, *existingTag)
			}
		}

		// Replace tags with processed ones
		post.Tags = processedTags

		// Update post and replace associations
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(post).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *blogRepository) DeleteWithTags(id uuid.UUID) error {
	// GORM akan otomatis delete associations karena ON DELETE CASCADE
	return r.db.Select("Tags").Delete(&model.BlogPost{ID: id}).Error
}

func (r *blogRepository) GetAllWithTags() ([]model.BlogPost, error) {
	var posts []model.BlogPost
	err := r.db.Preload("Tags").Order("created_at DESC").Find(&posts).Error
	return posts, err
}

func (r *blogRepository) GetPublishedWithTags() ([]model.BlogPost, error) {
	var posts []model.BlogPost
	err := r.db.Preload("Tags").
		Where("status = ?", "published").
		Order("publish_date DESC").
		Find(&posts).Error
	return posts, err
}

func (r *blogRepository) IncrementViewCount(id uuid.UUID) error {
	return r.db.Model(&model.BlogPost{}).
		Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *blogRepository) CreateTag(tag *model.BlogTag) error {
	return r.db.Create(tag).Error
}

func (r *blogRepository) GetOrCreateTag(name string) (*model.BlogTag, error) {
	return r.getOrCreateTagTx(r.db, name)
}

func (r *blogRepository) getOrCreateTagTx(tx *gorm.DB, name string) (*model.BlogTag, error) {
	var tag model.BlogTag
	err := tx.Where("name = ?", name).First(&tag).Error
	if err == gorm.ErrRecordNotFound {
		tag = model.BlogTag{Name: name}
		if err := tx.Create(&tag).Error; err != nil {
			return nil, err
		}
		return &tag, nil
	}
	return &tag, err
}

func (r *blogRepository) GetAllTags() ([]model.BlogTag, error) {
	var tags []model.BlogTag
	err := r.db.Order("name ASC").Find(&tags).Error
	return tags, err
}

// ============================
// SECTIONS REPOSITORY
// ============================

type SectionRepository interface {
	Create(section *model.Section) error
	Delete(id uuid.UUID) error
	GetAll() ([]model.Section, error)
}

type sectionRepository struct {
	db *gorm.DB
}

func NewSectionRepository(db *gorm.DB) SectionRepository {
	return &sectionRepository{db: db}
}

func (r *sectionRepository) Create(section *model.Section) error {
	return r.db.Create(section).Error
}

func (r *sectionRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Section{}).Error
}

func (r *sectionRepository) GetAll() ([]model.Section, error) {
	var sections []model.Section
	err := r.db.Order("display_order ASC").Find(&sections).Error
	return sections, err
}

// ============================
// SOCIAL LINKS REPOSITORY
// ============================

type SocialLinkRepository interface {
	Create(link *model.SocialLink) error
	Delete(id uuid.UUID) error
	GetAll() ([]model.SocialLink, error)
}

type socialLinkRepository struct {
	db *gorm.DB
}

func NewSocialLinkRepository(db *gorm.DB) SocialLinkRepository {
	return &socialLinkRepository{db: db}
}

func (r *socialLinkRepository) Create(link *model.SocialLink) error {
	return r.db.Create(link).Error
}

func (r *socialLinkRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.SocialLink{}).Error
}

func (r *socialLinkRepository) GetAll() ([]model.SocialLink, error) {
	var links []model.SocialLink
	err := r.db.Order("display_order ASC").Find(&links).Error
	return links, err
}

// ============================
// SETTINGS REPOSITORY
// ============================

type SettingRepository interface {
	Create(setting *model.Setting) error
	Delete(id uuid.UUID) error
	GetAll() ([]model.Setting, error)
}

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) Create(setting *model.Setting) error {
	return r.db.Create(setting).Error
}

func (r *settingRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Setting{}).Error
}

func (r *settingRepository) GetAll() ([]model.Setting, error) {
	var settings []model.Setting
	err := r.db.Order("key ASC").Find(&settings).Error
	return settings, err
}
