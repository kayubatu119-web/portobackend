-- +migrate Up
-- +migrate StatementBegin

-- =======================================
-- Timezone Indonesia (WIB)
-- =======================================

-- Aktifkan ekstensi UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set timezone ke Asia/Jakarta (WIB)
SET TIME ZONE 'Asia/Jakarta';

-- ============================
-- PORTFOLIO SECTIONS TABLE
-- ============================

CREATE TABLE portfolio_sections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id      VARCHAR(50) UNIQUE NOT NULL,
    label           VARCHAR(100) NOT NULL,
    display_order   INTEGER NOT NULL DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- PROJECTS TABLE
-- ============================

CREATE TABLE portfolio_projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           VARCHAR(200) NOT NULL,
    description     TEXT NOT NULL,
    image_url       VARCHAR(500),
    demo_url        VARCHAR(500),
    code_url        VARCHAR(500) NOT NULL,
    display_order   INTEGER NOT NULL DEFAULT 0,
    is_featured     BOOLEAN DEFAULT FALSE,
    status          VARCHAR(20) DEFAULT 'published', -- draft, published, archived
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- PROJECT TAGS TABLE (Many-to-Many)
-- ============================

CREATE TABLE project_tags (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(50) UNIQUE NOT NULL,
    color           VARCHAR(7), -- HEX color code
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE project_tag_relations (
    project_id      UUID REFERENCES portfolio_projects(id) ON DELETE CASCADE,
    tag_id          UUID REFERENCES project_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, tag_id)
);

-- ============================
-- EXPERIENCES TABLE
-- ============================

CREATE TABLE portfolio_experiences (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           VARCHAR(200) NOT NULL,
    company         VARCHAR(150) NOT NULL,
    location        VARCHAR(200),
    start_year      VARCHAR(20),
    end_year        VARCHAR(20),
    current_job     BOOLEAN DEFAULT FALSE,
    display_order   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- EXPERIENCE RESPONSIBILITIES TABLE
-- ============================

CREATE TABLE experience_responsibilities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experience_id   UUID REFERENCES portfolio_experiences(id) ON DELETE CASCADE,
    description     TEXT NOT NULL,
    display_order   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- EXPERIENCE SKILLS TABLE (Many-to-Many)
-- ============================

CREATE TABLE experience_skills (
    experience_id   UUID REFERENCES portfolio_experiences(id) ON DELETE CASCADE,
    skill_name      VARCHAR(100) NOT NULL,
    PRIMARY KEY (experience_id, skill_name)
);

-- ============================
-- SKILLS TABLE
-- ============================

CREATE TABLE portfolio_skills (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) UNIQUE NOT NULL,
    value           INTEGER CHECK (value >= 0 AND value <= 100),
    icon_url        VARCHAR(500),
    category        VARCHAR(50), -- programming, framework, tool, etc
    display_order   INTEGER NOT NULL DEFAULT 0,
    is_featured     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- CERTIFICATES TABLE
-- ============================

CREATE TABLE portfolio_certificates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(200) NOT NULL,
    image_url       VARCHAR(500) NOT NULL,
    issue_date      DATE,
    issuer          VARCHAR(150),
    credential_url  VARCHAR(500),
    display_order   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- EDUCATION TABLE
-- ============================

CREATE TABLE portfolio_education (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school          VARCHAR(200) NOT NULL,
    major           VARCHAR(200) NOT NULL,
    start_year      VARCHAR(20),
    end_year        VARCHAR(20),
    description     TEXT,
    degree          VARCHAR(100), -- S1, S2, SMA, etc
    display_order   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- EDUCATION ACHIEVEMENTS TABLE
-- ============================

CREATE TABLE education_achievements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    education_id    UUID REFERENCES portfolio_education(id) ON DELETE CASCADE,
    achievement     TEXT NOT NULL,
    display_order   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- TESTIMONIALS TABLE
-- ============================

CREATE TABLE portfolio_testimonials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    title           VARCHAR(150) NOT NULL,
    message         TEXT NOT NULL,
    avatar_url      VARCHAR(500),
    rating          INTEGER CHECK (rating >= 1 AND rating <= 5),
    is_featured     BOOLEAN DEFAULT FALSE,
    display_order   INTEGER NOT NULL DEFAULT 0,
    status          VARCHAR(20) DEFAULT 'approved', -- pending, approved, rejected
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- BLOG POSTS TABLE
-- ============================

CREATE TABLE portfolio_blog_posts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           VARCHAR(200) NOT NULL,
    content         TEXT,
    excerpt         TEXT,
    slug            VARCHAR(200) UNIQUE NOT NULL,
    featured_image  VARCHAR(500),
    publish_date    DATE,
    status          VARCHAR(20) DEFAULT 'draft', -- draft, published, archived
    view_count      INTEGER DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- BLOG TAGS TABLE (Many-to-Many)
-- ============================

CREATE TABLE blog_tags (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(50) UNIQUE NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE blog_post_tags (
    post_id         UUID REFERENCES portfolio_blog_posts(id) ON DELETE CASCADE,
    tag_id          UUID REFERENCES blog_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

-- ============================
-- SOCIAL LINKS TABLE
-- ============================

CREATE TABLE portfolio_social_links (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform        VARCHAR(50) UNIQUE NOT NULL, -- whatsapp, instagram, linkedin, etc
    url             VARCHAR(500) NOT NULL,
    icon_name       VARCHAR(50), -- react-icons name
    display_order   INTEGER NOT NULL DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- SITE SETTINGS TABLE
-- ============================

CREATE TABLE portfolio_settings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key             VARCHAR(100) UNIQUE NOT NULL,
    value           TEXT,
    data_type       VARCHAR(20) DEFAULT 'string', -- string, number, boolean, json
    description     TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- CONTACT MESSAGES TABLE
-- ============================

CREATE TABLE contact_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    email           VARCHAR(150) NOT NULL,
    subject         VARCHAR(200),
    message         TEXT NOT NULL,
    ip_address      INET,
    user_agent      TEXT,
    status          VARCHAR(20) DEFAULT 'unread', -- unread, read, replied, archived
    replied_at      TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================
-- INSERT DEFAULT DATA
-- ============================

-- Portfolio Sections
INSERT INTO portfolio_sections (section_id, label, display_order) VALUES
('profil', 'Profil', 1),
('about', 'About', 2),
('projek', 'Projek', 3),
('pengalaman', 'Pengalaman', 4),
('skill', 'Skill', 5),
('studi', 'Studi', 6),
('testimoni', 'Testimoni', 7),
('blog', 'Blog', 8),
('kontak', 'Kontak', 9);

-- Project Tags
INSERT INTO project_tags (name, color) VALUES
('Laravel', '#FF2D20'),
('Bootstrap', '#7952B3'),
('Node.js', '#339933'),
('MySQL', '#4479A1'),
('Tailwind', '#06B6D4'),
('Python', '#3776AB'),
('Face Recognition', '#FFD43B'),
('Filament', '#F59E0B'),
('API', '#10B981'),
('AI', '#8B5CF6'),
('React', '#61DAFB'),
('JavaScript', '#F7DF1E');

-- Social Links
INSERT INTO portfolio_social_links (platform, url, icon_name, display_order) VALUES
('WhatsApp', 'https://wa.me/6285809735614', 'FaWhatsapp', 1),
('Instagram', 'https://instagram.com/mfathir_fh', 'FiInstagram', 2),
('LinkedIn', 'https://linkedin.com/in/muhammad-fathiir-farhansyah-58baa6279', 'FiLinkedin', 3),
('Telegram', 'https://t.me/Mffathir', 'FaTelegram', 4),
('Line', 'https://line.me/ti/p/5JxYtPuxe3', 'FaLine', 5),
('Email', 'mailto:fathirfarhansyah24@gmail.com', 'FiMail', 6),
('GitHub', 'https://github.com/mffathir-24', 'FiGithub', 7);

-- Site Settings
INSERT INTO portfolio_settings (key, value, data_type, description) VALUES
('site_title', 'Muhammad Fathiir Farhansyah - Portfolio', 'string', 'Judul website portfolio'),
('site_description', 'Portfolio of Muhammad Fathiir Farhansyah - Web Developer | Laravel Junior | FullStack', 'string', 'Deskripsi meta website'),
('contact_email', 'fathirfarhansyah24@gmail.com', 'string', 'Email utama untuk kontak'),
('phone_number', '+6285809735614', 'string', 'Nomor telepon'),
('location', 'Kalidoni, Palembang, Sumatera Selatan, Indonesia', 'string', 'Lokasi'),
('cv_url', '/CV-Muhammad-Fathiir-Farhansyah.pdf', 'string', 'URL file CV'),
('theme', 'dark', 'string', 'Tema website');

-- +migrate StatementEnd