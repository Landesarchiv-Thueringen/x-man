package db

import (
	"time"
)

type RecordObjectAppraisal struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Code      string    `gorm:"unique" xml:"code" json:"code"`
	ShortDesc string    `json:"shortDesc"`
	Desc      string    `json:"desc"`
}

type ConfidentialityLevel struct {
	ID        string    `gorm:"primaryKey" json:"code"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	ShortDesc string    `json:"shortDesc"`
	Desc      string    `json:"desc"`
}

type Medium struct {
	ID        string    `gorm:"primaryKey" json:"code"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	ShortDesc string    `json:"shortDesc"`
	Desc      string    `json:"desc"`
}
