package models

import "gorm.io/gorm"

type BlobUpload struct {
	gorm.Model
	ID         uint `gorm:"primaryKey"`
	PartNumber int
	UploadID   string
	Etag       string
	Repository string
	FileID     string
	Size       int64
}
