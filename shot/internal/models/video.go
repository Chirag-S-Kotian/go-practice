package models

import (
	"time"
)

type Video struct {
	ID          string
	Title       string
	Description string
	VideoURL    string
	UserID      int
	CreatedAt   time.Time
}

func CreateVideo(title, desc, url string, userID int) error {
	_, err := DB.Exec(`
		INSERT INTO videos (title, description, video_url, user_id)
		VALUES ($1, $2, $3, $4)`,
		title, desc, url, userID)
	return err
}
