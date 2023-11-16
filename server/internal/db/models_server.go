package db

type ServerState struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	DatabaseInit bool `gorm:"default:true" json:"databaseInit"`
}
