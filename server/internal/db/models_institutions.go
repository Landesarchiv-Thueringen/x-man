package db

import (
	"database/sql/driver"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ConfiguredInstitution represents an institution as configured in the
// administration panel.
//
// It maps a transfer directory to assigned users and an archive collection.
//
// All messages that are received via the configured transfer directory are
// considered to belong the the configured institution, ignoring the content of
// the "sender" field.
type ConfiguredInstitution struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         string         `json:"name"`
	Abbreviation string         `json:"abbreviation"`
	// The assigned users' objectGUIDs as by LDAP
	UserIDs           pq.ByteaArray     `gorm:"type:bytea[];not null" json:"userIds"`
	CollectionID      *uint             `json:"collectionId"`
	Collection        *Collection       `json:"collection"`
	TransferDirectory TransferDirectory `gorm:"foreignKey:InstitutionID" json:"transferDirectory"`
}

type TransferDirectory struct {
	ID            uint           `gorm:"primaryKey" json:"-"`
	CreatedAt     time.Time      `json:"-"`
	UpdatedAt     time.Time      `json:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	InstitutionID uint           `json:"-"`
	URI           string         `json:"uri"`
	Username      string         `json:"username"`
	Password      string         `json:"password"`
}

type Collection struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `json:"name"`
}

type StringArray []string

func (array *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("value cannot cast to []byte")
	}
	*array = strings.Split(string(bytes), ",")
	return nil
}

func (array StringArray) Value() (driver.Value, error) {
	if len(array) == 0 {
		return nil, nil
	}
	return strings.Join(array, ","), nil
}
