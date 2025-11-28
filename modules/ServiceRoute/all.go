package serviceroute

import (
	"gintugas/modules/components/all/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ============================
// SKILLS HANDLER
// ============================

type SkillHandler struct {
	service service.SkillService
}

func NewSkillHandler(service service.SkillService) *SkillHandler {
	return &SkillHandler{service: service}
}

func (h *SkillHandler) Create(c *gin.Context) {
	skill, err := h.service.Create(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Skill created successfully",
		"data":    skill,
	})
}

func (h *SkillHandler) GetByID(c *gin.Context) {
	skill, err := h.service.GetByID(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Skill retrieved successfully",
		"data":    skill,
	})
}

func (h *SkillHandler) Update(c *gin.Context) {
	skill, err := h.service.Update(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Skill updated successfully",
		"data":    skill,
	})
}

func (h *SkillHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Skill deleted successfully",
	})
}

func (h *SkillHandler) GetAll(c *gin.Context) {
	skills, err := h.service.GetAll(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Skills retrieved successfully",
		"data":    skills,
	})
}

func (h *SkillHandler) GetFeatured(c *gin.Context) {
	skills, err := h.service.GetFeatured(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Featured skills retrieved successfully",
		"data":    skills,
	})
}

func (h *SkillHandler) GetByCategory(c *gin.Context) {
	skills, err := h.service.GetByCategory(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Skills by category retrieved successfully",
		"data":    skills,
	})
}

func (h *SkillHandler) CreateWithIcon(c *gin.Context) {
	response, err := h.service.CreateWithIcon(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    response,
		"message": "Skill created successfully with icon upload",
	})
}

func (h *SkillHandler) UpdateWithIcon(c *gin.Context) {
	response, err := h.service.UpdateWithIcon(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"message": "Skill updated successfully with icon upload",
	})
}

// ============================
// CERTIFICATES HANDLER
// ============================

type CertificateHandler struct {
	service service.CertificateService
}

func NewCertificateHandler(service service.CertificateService) *CertificateHandler {
	return &CertificateHandler{service: service}
}

func (h *CertificateHandler) Create(c *gin.Context) {
	cert, err := h.service.Create(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Certificate created successfully",
		"data":    cert,
	})
}

func (h *CertificateHandler) GetByID(c *gin.Context) {
	cert, err := h.service.GetByID(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate retrieved successfully",
		"data":    cert,
	})
}

func (h *CertificateHandler) Update(c *gin.Context) {
	cert, err := h.service.Update(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate updated successfully",
		"data":    cert,
	})
}

func (h *CertificateHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate deleted successfully",
	})
}

func (h *CertificateHandler) GetAll(c *gin.Context) {
	certs, err := h.service.GetAll(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificates retrieved successfully",
		"data":    certs,
	})
}

func (h *CertificateHandler) CreateWithImage(c *gin.Context) {
	response, err := h.service.CreateWithImage(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    response,
		"message": "Certificate created successfully with image upload",
	})
}

// ============================
// EDUCATION HANDLER
// ============================

type EducationHandler struct {
	service service.EducationService
}

func NewEducationHandler(service service.EducationService) *EducationHandler {
	return &EducationHandler{service: service}
}

func (h *EducationHandler) CreateWithAchievements(c *gin.Context) {
	edu, err := h.service.CreateWithAchievements(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Education created successfully",
		"data":    edu,
	})
}

func (h *EducationHandler) GetByIDWithAchievements(c *gin.Context) {
	edu, err := h.service.GetByIDWithAchievements(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Education retrieved successfully",
		"data":    edu,
	})
}

func (h *EducationHandler) UpdateWithAchievements(c *gin.Context) {
	edu, err := h.service.UpdateWithAchievements(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Education updated successfully",
		"data":    edu,
	})
}

func (h *EducationHandler) DeleteWithAchievements(c *gin.Context) {
	if err := h.service.DeleteWithAchievements(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Education deleted successfully",
	})
}

func (h *EducationHandler) GetAllWithAchievements(c *gin.Context) {
	educations, err := h.service.GetAllWithAchievements(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Educations retrieved successfully",
		"data":    educations,
	})
}

// ============================
// TESTIMONIALS HANDLER
// ============================

type TestimonialHandler struct {
	service service.TestimonialService
}

func NewTestimonialHandler(service service.TestimonialService) *TestimonialHandler {
	return &TestimonialHandler{service: service}
}

func (h *TestimonialHandler) Create(c *gin.Context) {
	test, err := h.service.Create(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Testimonial created successfully",
		"data":    test,
	})
}

func (h *TestimonialHandler) GetByID(c *gin.Context) {
	test, err := h.service.GetByID(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Testimonial retrieved successfully",
		"data":    test,
	})
}

func (h *TestimonialHandler) Update(c *gin.Context) {
	test, err := h.service.Update(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Testimonial updated successfully",
		"data":    test,
	})
}

func (h *TestimonialHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Testimonial deleted successfully",
	})
}

func (h *TestimonialHandler) GetAll(c *gin.Context) {
	testimonials, err := h.service.GetAll(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Testimonials retrieved successfully",
		"data":    testimonials,
	})
}

func (h *TestimonialHandler) GetFeatured(c *gin.Context) {
	testimonials, err := h.service.GetFeatured(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Featured testimonials retrieved successfully",
		"data":    testimonials,
	})
}

func (h *TestimonialHandler) GetByStatus(c *gin.Context) {
	testimonials, err := h.service.GetByStatus(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Testimonials by status retrieved successfully",
		"data":    testimonials,
	})
}

// ============================
// BLOG HANDLER
// ============================

type BlogHandler struct {
	service service.BlogService
}

func NewBlogHandler(service service.BlogService) *BlogHandler {
	return &BlogHandler{service: service}
}

func (h *BlogHandler) CreateWithTags(c *gin.Context) {
	post, err := h.service.CreateWithTags(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Blog post created successfully",
		"data":    post,
	})
}

func (h *BlogHandler) GetByIDWithTags(c *gin.Context) {
	post, err := h.service.GetByIDWithTags(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog post retrieved successfully",
		"data":    post,
	})
}

func (h *BlogHandler) GetBySlugWithTags(c *gin.Context) {
	post, err := h.service.GetBySlugWithTags(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog post retrieved successfully",
		"data":    post,
	})
}

func (h *BlogHandler) UpdateWithTags(c *gin.Context) {
	post, err := h.service.UpdateWithTags(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog post updated successfully",
		"data":    post,
	})
}

func (h *BlogHandler) DeleteWithTags(c *gin.Context) {
	if err := h.service.DeleteWithTags(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog post deleted successfully",
	})
}

func (h *BlogHandler) GetAllWithTags(c *gin.Context) {
	posts, err := h.service.GetAllWithTags(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog posts retrieved successfully",
		"data":    posts,
	})
}

func (h *BlogHandler) GetPublishedWithTags(c *gin.Context) {
	posts, err := h.service.GetPublishedWithTags(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Published blog posts retrieved successfully",
		"data":    posts,
	})
}

func (h *BlogHandler) GetAllTags(c *gin.Context) {
	tags, err := h.service.GetAllTags(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tags retrieved successfully",
		"data":    tags,
	})
}

// ============================
// SECTIONS HANDLER
// ============================

type SectionHandler struct {
	service service.SectionService
}

func NewSectionHandler(service service.SectionService) *SectionHandler {
	return &SectionHandler{service: service}
}

func (h *SectionHandler) Create(c *gin.Context) {
	section, err := h.service.Create(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Section created successfully",
		"data":    section,
	})
}

func (h *SectionHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Section deleted successfully",
	})
}

func (h *SectionHandler) GetAll(c *gin.Context) {
	sections, err := h.service.GetAll(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sections retrieved successfully",
		"data":    sections,
	})
}

// ============================
// SOCIAL LINKS HANDLER
// ============================

type SocialLinkHandler struct {
	service service.SocialLinkService
}

func NewSocialLinkHandler(service service.SocialLinkService) *SocialLinkHandler {
	return &SocialLinkHandler{service: service}
}

func (h *SocialLinkHandler) Create(c *gin.Context) {
	link, err := h.service.Create(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Social link created successfully",
		"data":    link,
	})
}

func (h *SocialLinkHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Social link deleted successfully",
	})
}

func (h *SocialLinkHandler) GetAll(c *gin.Context) {
	links, err := h.service.GetAll(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Social links retrieved successfully",
		"data":    links,
	})
}

// ============================
// SETTINGS HANDLER
// ============================

type SettingHandler struct {
	service service.SettingService
}

func NewSettingHandler(service service.SettingService) *SettingHandler {
	return &SettingHandler{service: service}
}

func (h *SettingHandler) Create(c *gin.Context) {
	setting, err := h.service.Create(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Setting created successfully",
		"data":    setting,
	})
}

func (h *SettingHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting deleted successfully",
	})
}

func (h *SettingHandler) GetAll(c *gin.Context) {
	settings, err := h.service.GetAll(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Settings retrieved successfully",
		"data":    settings,
	})
}
