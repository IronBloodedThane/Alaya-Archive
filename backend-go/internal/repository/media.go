package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Media struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	MediaType       string    `json:"media_type"`
	Title           string    `json:"title"`
	TitleOriginal   string    `json:"title_original,omitempty"`
	Description     string    `json:"description,omitempty"`
	CoverImage      string    `json:"cover_image,omitempty"`
	Status          string    `json:"status"`
	Rating          *int      `json:"rating"`
	Notes           string    `json:"notes,omitempty"`
	YearReleased    *int      `json:"year_released"`
	Creator         string    `json:"creator,omitempty"`
	Genre           string    `json:"genre,omitempty"`
	VolumesTotal    *int      `json:"volumes_total"`
	VolumesOwned    *int      `json:"volumes_owned"`
	EpisodesTotal   *int      `json:"episodes_total"`
	EpisodesWatched *int      `json:"episodes_watched"`
	ChaptersTotal   *int      `json:"chapters_total"`
	ChaptersRead    *int      `json:"chapters_read"`
	ISBN            string    `json:"isbn,omitempty"`
	Series          string    `json:"series,omitempty"`
	SeriesPosition  *int      `json:"series_position,omitempty"`
	ListType        string    `json:"list_type"`
	IsPublic        bool      `json:"is_public"`
	Tags            []string  `json:"tags,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type MediaFilter struct {
	MediaType string
	Status    string
	Search    string
	Tag       string
	ListType  string
	Series    string
	Limit     int
	Offset    int
}

type MediaRepository struct {
	db *sql.DB
}

func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(m *Media) error {
	if m.ListType == "" {
		m.ListType = "owned"
	}
	_, err := r.db.Exec(
		`INSERT INTO media (id, user_id, media_type, title, title_original, description, cover_image, status, rating, notes, year_released, creator, genre, volumes_total, volumes_owned, episodes_total, episodes_watched, chapters_total, chapters_read, isbn, series, series_position, list_type, is_public)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.UserID, m.MediaType, m.Title, m.TitleOriginal, m.Description, m.CoverImage, m.Status, m.Rating, m.Notes, m.YearReleased, m.Creator, m.Genre, m.VolumesTotal, m.VolumesOwned, m.EpisodesTotal, m.EpisodesWatched, m.ChaptersTotal, m.ChaptersRead, m.ISBN, m.Series, m.SeriesPosition, m.ListType, m.IsPublic,
	)
	return err
}

// FindDuplicates returns media rows belonging to userID that share the same
// (mediaType, isbn) as the caller. Empty isbn returns no matches — we don't
// treat blank ISBNs as a fingerprint.
func (r *MediaRepository) FindDuplicates(userID, mediaType, isbn string) ([]*Media, error) {
	if isbn == "" {
		return nil, nil
	}
	rows, err := r.db.Query(
		`SELECT id, user_id, media_type, title, title_original, description, cover_image, status, rating, notes, year_released, creator, genre, volumes_total, volumes_owned, episodes_total, episodes_watched, chapters_total, chapters_read, isbn, series, series_position, list_type, is_public, created_at, updated_at
		 FROM media WHERE user_id = ? AND media_type = ? AND isbn = ? ORDER BY created_at ASC`,
		userID, mediaType, isbn,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Media
	for rows.Next() {
		m := &Media{}
		if err := rows.Scan(&m.ID, &m.UserID, &m.MediaType, &m.Title, &m.TitleOriginal, &m.Description, &m.CoverImage, &m.Status, &m.Rating, &m.Notes, &m.YearReleased, &m.Creator, &m.Genre, &m.VolumesTotal, &m.VolumesOwned, &m.EpisodesTotal, &m.EpisodesWatched, &m.ChaptersTotal, &m.ChaptersRead, &m.ISBN, &m.Series, &m.SeriesPosition, &m.ListType, &m.IsPublic, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	for _, m := range items {
		tags, _ := r.GetTags(m.ID)
		m.Tags = tags
	}
	return items, nil
}

func (r *MediaRepository) GetByID(id string) (*Media, error) {
	m := &Media{}
	err := r.db.QueryRow(
		`SELECT id, user_id, media_type, title, title_original, description, cover_image, status, rating, notes, year_released, creator, genre, volumes_total, volumes_owned, episodes_total, episodes_watched, chapters_total, chapters_read, isbn, series, series_position, list_type, is_public, created_at, updated_at
		 FROM media WHERE id = ?`, id,
	).Scan(&m.ID, &m.UserID, &m.MediaType, &m.Title, &m.TitleOriginal, &m.Description, &m.CoverImage, &m.Status, &m.Rating, &m.Notes, &m.YearReleased, &m.Creator, &m.Genre, &m.VolumesTotal, &m.VolumesOwned, &m.EpisodesTotal, &m.EpisodesWatched, &m.ChaptersTotal, &m.ChaptersRead, &m.ISBN, &m.Series, &m.SeriesPosition, &m.ListType, &m.IsPublic, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}

	tags, _ := r.GetTags(id)
	m.Tags = tags
	return m, nil
}

func (r *MediaRepository) List(userID string, filter MediaFilter) ([]*Media, int, error) {
	where := []string{"m.user_id = ?"}
	args := []interface{}{userID}

	if filter.MediaType != "" {
		where = append(where, "m.media_type = ?")
		args = append(args, filter.MediaType)
	}
	if filter.Status != "" {
		where = append(where, "m.status = ?")
		args = append(args, filter.Status)
	}
	if filter.ListType != "" {
		where = append(where, "m.list_type = ?")
		args = append(args, filter.ListType)
	}
	if filter.Series != "" {
		where = append(where, "m.series = ?")
		args = append(args, filter.Series)
	}
	if filter.Search != "" {
		where = append(where, "(m.title LIKE ? OR m.creator LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s)
	}
	if filter.Tag != "" {
		where = append(where, "m.id IN (SELECT media_id FROM media_tags mt JOIN tags t ON mt.tag_id = t.id WHERE t.name = ?)")
		args = append(args, filter.Tag)
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM media m WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(
		`SELECT m.id, m.user_id, m.media_type, m.title, m.title_original, m.description, m.cover_image, m.status, m.rating, m.notes, m.year_released, m.creator, m.genre, m.volumes_total, m.volumes_owned, m.episodes_total, m.episodes_watched, m.chapters_total, m.chapters_read, m.isbn, m.series, m.series_position, m.list_type, m.is_public, m.created_at, m.updated_at
		 FROM media m WHERE %s ORDER BY m.updated_at DESC LIMIT ? OFFSET ?`, whereClause,
	)
	args = append(args, limit, filter.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}

	var items []*Media
	for rows.Next() {
		m := &Media{}
		if err := rows.Scan(&m.ID, &m.UserID, &m.MediaType, &m.Title, &m.TitleOriginal, &m.Description, &m.CoverImage, &m.Status, &m.Rating, &m.Notes, &m.YearReleased, &m.Creator, &m.Genre, &m.VolumesTotal, &m.VolumesOwned, &m.EpisodesTotal, &m.EpisodesWatched, &m.ChaptersTotal, &m.ChaptersRead, &m.ISBN, &m.Series, &m.SeriesPosition, &m.ListType, &m.IsPublic, &m.CreatedAt, &m.UpdatedAt); err != nil {
			rows.Close()
			return nil, 0, err
		}
		items = append(items, m)
	}
	rows.Close()

	for _, m := range items {
		tags, _ := r.GetTags(m.ID)
		m.Tags = tags
	}

	return items, total, nil
}

func (r *MediaRepository) ListPublicByUser(userID string, filter MediaFilter) ([]*Media, int, error) {
	where := []string{"m.user_id = ?", "m.is_public = 1"}
	args := []interface{}{userID}

	if filter.MediaType != "" {
		where = append(where, "m.media_type = ?")
		args = append(args, filter.MediaType)
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM media m WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(
		`SELECT m.id, m.user_id, m.media_type, m.title, m.title_original, m.description, m.cover_image, m.status, m.rating, m.notes, m.year_released, m.creator, m.genre, m.volumes_total, m.volumes_owned, m.episodes_total, m.episodes_watched, m.chapters_total, m.chapters_read, m.isbn, m.series, m.series_position, m.list_type, m.is_public, m.created_at, m.updated_at
		 FROM media m WHERE %s ORDER BY m.updated_at DESC LIMIT ? OFFSET ?`, whereClause,
	)
	args = append(args, limit, filter.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*Media
	for rows.Next() {
		m := &Media{}
		if err := rows.Scan(&m.ID, &m.UserID, &m.MediaType, &m.Title, &m.TitleOriginal, &m.Description, &m.CoverImage, &m.Status, &m.Rating, &m.Notes, &m.YearReleased, &m.Creator, &m.Genre, &m.VolumesTotal, &m.VolumesOwned, &m.EpisodesTotal, &m.EpisodesWatched, &m.ChaptersTotal, &m.ChaptersRead, &m.ISBN, &m.Series, &m.SeriesPosition, &m.ListType, &m.IsPublic, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}

	return items, total, nil
}

func (r *MediaRepository) Update(m *Media) error {
	if m.ListType == "" {
		m.ListType = "owned"
	}
	_, err := r.db.Exec(
		`UPDATE media SET title = ?, title_original = ?, description = ?, cover_image = ?, status = ?, rating = ?, notes = ?, year_released = ?, creator = ?, genre = ?, volumes_total = ?, volumes_owned = ?, episodes_total = ?, episodes_watched = ?, chapters_total = ?, chapters_read = ?, isbn = ?, series = ?, series_position = ?, list_type = ?, is_public = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		m.Title, m.TitleOriginal, m.Description, m.CoverImage, m.Status, m.Rating, m.Notes, m.YearReleased, m.Creator, m.Genre, m.VolumesTotal, m.VolumesOwned, m.EpisodesTotal, m.EpisodesWatched, m.ChaptersTotal, m.ChaptersRead, m.ISBN, m.Series, m.SeriesPosition, m.ListType, m.IsPublic, m.ID,
	)
	return err
}

func (r *MediaRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM media WHERE id = ?`, id)
	return err
}

func (r *MediaRepository) GetStats(userID string) (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	var total int
	r.db.QueryRow(`SELECT COUNT(*) FROM media WHERE user_id = ?`, userID).Scan(&total)
	stats["total"] = total

	rows, err := r.db.Query(`SELECT media_type, COUNT(*) as count FROM media WHERE user_id = ? GROUP BY media_type`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byType := map[string]int{}
	for rows.Next() {
		var t string
		var c int
		rows.Scan(&t, &c)
		byType[t] = c
	}
	stats["by_type"] = byType

	rows2, err := r.db.Query(`SELECT status, COUNT(*) as count FROM media WHERE user_id = ? GROUP BY status`, userID)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	byStatus := map[string]int{}
	for rows2.Next() {
		var s string
		var c int
		rows2.Scan(&s, &c)
		byStatus[s] = c
	}
	stats["by_status"] = byStatus

	return stats, nil
}

func (r *MediaRepository) GetTags(mediaID string) ([]string, error) {
	rows, err := r.db.Query(`SELECT t.name FROM tags t JOIN media_tags mt ON t.id = mt.tag_id WHERE mt.media_id = ?`, mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tags = append(tags, name)
	}
	return tags, nil
}

func (r *MediaRepository) SetTags(mediaID string, tagNames []string) error {
	r.db.Exec(`DELETE FROM media_tags WHERE media_id = ?`, mediaID)

	for _, name := range tagNames {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" {
			continue
		}
		var tagID string
		err := r.db.QueryRow(`SELECT id FROM tags WHERE name = ?`, name).Scan(&tagID)
		if err != nil {
			tagID = generateID()
			r.db.Exec(`INSERT INTO tags (id, name) VALUES (?, ?)`, tagID, name)
		}
		r.db.Exec(`INSERT OR IGNORE INTO media_tags (media_id, tag_id) VALUES (?, ?)`, mediaID, tagID)
	}
	return nil
}

func generateID() string {
	// Simple UUID-like ID generation
	b := make([]byte, 16)
	// Use crypto/rand in production
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
