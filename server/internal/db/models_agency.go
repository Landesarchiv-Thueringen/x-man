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
	// Prefix is the agency prefix as by xdomea.
	Prefix string `json:"prefix"`
	// Code is the agency code as by xdomea.
	Code string `json:"code"`
	// ContactEmail is the e-mail address to use to contact the agency.
	ContactEmail string `json:"contactEmail"`
	// TransferDir is the directory that the Agency uses to pass messages.
	TransferDir           string                 `json:"transferDir"`
	FilesystemTransferDir *FilesystemTransferDir `json:"filesystemTransferDir"`
	WebDAVTransferDir     *WebDAVTransferDir     `json:"webDAVTransferDir"`
	// Users are users responsible for processes of this Agency.
	Users        []User      `gorm:"many2many:agency_users" json:"users"`
	CollectionID *uint       `json:"collectionId"`
	Collection   *Collection `json:"collection"`
}

// Contains all information needed to establish an directory of a filesystem as transfer directory.
type FilesystemTransferDir struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Path      string    `json:"name"`
	AgencyID  uint      `json:"-"`
}

// Contains all information needed to establish an directory of a filesystem as transfer directory.
type WebDAVTransferDir struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	URL       string    `json:"url"`
	User      *string   `json:"user"`
	Password  *string   `json:"password"`
	AgencyID  uint      `json:"-"`
}

type User struct {
	// ID is the LDAP objectGUID
	ID        []byte    `gorm:"primaryKey;type:bytea" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	// Agencies is the list of agencies that the user is responsible for.
	Agencies []Agency `gorm:"many2many:agency_users" json:"agencies"`
	// Preferences are settings the user can choose.
	Preferences UserPreferences `json:"preferences"`
}

type UserPreferences struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	UserID    []byte    `gorm:"type:bytea" json:"-"`
	// MessageEmailNotifications is the user's setting to receive e-mail notifications
	// regarding new messages from x-man.
	MessageEmailNotifications bool `json:"messageEmailNotifications"`
	// ErrorEmailNotifications is a setting for users with administration
	// privileges to receive e-mails for new processing errors.
	ErrorEmailNotifications bool `json:"errorEmailNotifications"`
}

func GetDefaultPreferences() UserPreferences {
	return UserPreferences{
		MessageEmailNotifications: true,
		ErrorEmailNotifications:   false,
	}
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
