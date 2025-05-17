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
func FetchAllVideos() ([]Video, error) {
	rows, err := DB.Query("SELECT id, title, description, video_url, user_id, created_at FROM videos ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.VideoURL, &v.UserID, &v.CreatedAt); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}

func FetchVideosByUser(userID int) ([]Video, error) {
	rows, err := DB.Query("SELECT id, title, description, video_url, user_id, created_at FROM videos WHERE user_id=$1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.VideoURL, &v.UserID, &v.CreatedAt); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}

func FetchVideoByID(videoID string) (Video, error) {
	var v Video
	err := DB.QueryRow("SELECT id, title, description, video_url, user_id, created_at FROM videos WHERE id=$1", videoID).
		Scan(&v.ID, &v.Title, &v.Description, &v.VideoURL, &v.UserID, &v.CreatedAt)
	return v, err
}
