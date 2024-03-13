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

func (RecordObjectAppraisal) TableName() string {
	return "appraisal_codelist"
}

type ConfidentialityLevel struct {
	ID        string    `gorm:"primaryKey" json:"code"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	ShortDesc string    `json:"shortDesc"`
	Desc      string    `json:"desc"`
}

func (ConfidentialityLevel) TableName() string {
	return "confidentiality_level_codelist"
}

type Medium struct {
	ID        string    `gorm:"primaryKey" json:"code"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	ShortDesc string    `json:"shortDesc"`
	Desc      string    `json:"desc"`
}

func (Medium) TableName() string {
	return "medium_codelist"
}

type MessageType struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Code      string    `json:"code"`
}

func (MessageType) TableName() string {
	return "message_type_codelist"
}
