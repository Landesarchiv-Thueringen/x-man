package app

import "lath/xman/internal/db"

const (
	XMAN_VERSION    = "v1.4.0"
	DefaultResponse = "x-man server " + XMAN_VERSION + " is running"
)

func Init() {
	db.Init()
}
