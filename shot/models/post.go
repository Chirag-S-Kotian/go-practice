package models

import (
	"github.com/docker/docker/api/types/time"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type Post struct {
	Title string
	userID uint
	createdAt 
}