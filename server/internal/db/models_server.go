package db

type ServerState struct {
	ID                uint `gorm:"primaryKey" json:"id"`
	MigrationComplete bool `gorm:"default:true" json:"databaseInit"`
}
