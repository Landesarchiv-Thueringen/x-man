package db

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Agency represents an institution as configured in the administration panel.
//
// It maps a transfer directory to assigned users and an archive collection.
//
// All messages that are received via the configured transfer directory are
// considered to belong the the configured institution, ignoring the content of
// the "sender" field.
type Agency struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         string         `json:"name"`
	Abbreviation string         `json:"abbreviation"`
	// TransferDir is the directory that the Agency uses to pass messages
	TransferDir string `json:"transferDir"`
	// UserIDs is the LDAP objectGUIDs of users responsible for processes of
	// this Agency
	UserIDs      pq.ByteaArray `gorm:"type:bytea[];not null" json:"userIds"`
	CollectionID *uint         `json:"collectionId"`
	Collection   *Collection   `json:"collection"`
}

// Collection refers to an archive collection within DIMAG.
//
// TODO: Retrieve collection information directly from DIMAG and remove this
// database table.
type Collection struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `json:"name"`
}
