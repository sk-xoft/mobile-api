package content

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("content not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Forms(ctx context.Context) ([]Form, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT FORM_ID, FORM_NAME, FORM_FILE_ID, FORM_FILE_NAME FROM m_forms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Form, 0)
	for rows.Next() {
		var item Form
		var id, name, fileID, fileName sql.NullString
		if err := rows.Scan(&id, &name, &fileID, &fileName); err != nil {
			return nil, err
		}
		item.FormID, item.FormName, item.FormFileID, item.FormFileName = id.String, name.String, fileID.String, fileName.String
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Form(ctx context.Context, id string) (Form, error) {
	var item Form
	var formID, name, fileID, fileName sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT FORM_ID, FORM_NAME, FORM_FILE_ID, FORM_FILE_NAME FROM m_forms WHERE FORM_ID=?", id).
		Scan(&formID, &name, &fileID, &fileName)
	if errors.Is(err, sql.ErrNoRows) {
		return Form{}, ErrNotFound
	}
	item.FormID, item.FormName, item.FormFileID, item.FormFileName = formID.String, name.String, fileID.String, fileName.String
	return item, err
}

func (r *Repository) News(ctx context.Context, kind string) ([]News, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT NEWS_ID, NEWS_TITLE, UPDATE_DATE FROM m_news WHERE NEWS_STATUS='A' AND NEWS_TYPE=? ORDER BY UPDATE_DATE DESC", kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]News, 0)
	for rows.Next() {
		var id int
		var title sql.NullString
		var date sql.NullTime
		if err := rows.Scan(&id, &title, &date); err != nil {
			return nil, err
		}
		newsDate := ""
		if date.Valid {
			newsDate = date.Time.Format("2006-01-02 15:04:05")
		}
		result = append(result, News{ID: fmt.Sprint(id), Title: title.String, NewsDate: newsDate})
	}
	return result, rows.Err()
}

func (r *Repository) NewsByID(ctx context.Context, id string) (News, string, error) {
	var item News
	var number int
	var file sql.NullString
	var title, body sql.NullString
	var date sql.NullTime
	err := r.db.QueryRowContext(ctx, "SELECT NEWS_ID, NEWS_TITLE, NEWS_DETAILS, NEWS_PICTURE_FILE_NAME, UPDATE_DATE FROM m_news WHERE NEWS_ID=?", id).
		Scan(&number, &title, &body, &file, &date)
	if errors.Is(err, sql.ErrNoRows) {
		return News{}, "", ErrNotFound
	}
	item.ID, item.Title, item.Body = fmt.Sprint(number), title.String, body.String
	if date.Valid {
		item.NewsDate = date.Time.Format("2006-01-02 15:04:05")
	}
	return item, file.String, err
}
