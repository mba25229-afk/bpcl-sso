package model

import (
	"time"

	"github.com/google/uuid"
)

type UploadedFile struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	OriginalName string     `json:"original_name"`
	StoredPath   string     `json:"stored_path"`
	SizeBytes    *int64     `json:"size_bytes"`
	MimeType     *string    `json:"mime_type"`
	UploadType   string     `json:"upload_type"`
	Status       string     `json:"status"`
	ErrorMessage *string    `json:"error_message"`
	RowCount     *int       `json:"row_count"`
	CCNumber     *string    `json:"cc_number"`
	Period       *time.Time `json:"period"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
