package db

import "time"

type ServerState struct {
	ID               uint      `gorm:"primaryKey" json:"-"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
	XManMajorVersion uint
	XManMinorVersion uint
	XManPatchVersion uint
}
