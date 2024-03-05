package db

import (
	"time"
)

// Agency represents an institution as configured in the administration panel.
//
// It maps a transfer directory to assigned users and an archive collection.
//
// All messages that are received via the configured transfer directory are
// considered to belong the the configured institution, ignoring the content of
// the "sender" field.
type Agency struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	Name         string    `json:"name"`
	Abbreviation string    `json:"abbreviation"`
	// Prefix is the agency prefix as by xdomea
	Prefix string `json:"prefix"`
	// Code is the agency code as by xdomea
	Code string `json:"code"`
	// TransferDir is the directory that the Agency uses to pass messages
	TransferDir string `json:"transferDir"`
	// Users are users responsible for processes of this Agency
	Users        []User      `gorm:"many2many:agency_users" json:"users"`
	CollectionID *uint       `json:"collectionId"`
	Collection   *Collection `json:"collection"`
}

type User struct {
	// ID is the LDAP objectGUID
	ID []byte `gorm:"primaryKey;type:bytea" json:"id"`
	// Agencies is the list of agencies that the user is responsible for
	Agencies []Agency `gorm:"many2many:agency_users" json:"agencies"`
	// EmailNotifications is the user's setting to receive e-mail notifications
	// from x-man
	EmailNotifications bool `json:"emailNotifications"`
}

// Collection refers to an archive collection within DIMAG.
//
// TODO: Retrieve collection information directly from DIMAG and remove this
// database table.
type Collection struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
}
