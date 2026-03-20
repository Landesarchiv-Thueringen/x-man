package app

import (
	"errors"
	"lath/xman/internal/db"
	"log"
)

const (
	XMAN_VERSION    = "v1.4.0"
	DefaultResponse = "x-man server " + XMAN_VERSION + " is running"
)

var ErrAppInit = errors.New("application initialization failed")

func Init() error {
	err := db.Init()
	if err != nil {
		log.Println(err)
		return ErrAppInit
	}
	return nil
}
