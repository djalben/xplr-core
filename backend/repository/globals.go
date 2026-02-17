package repository

import (
	"database/sql"
)

// GlobalDB - единая точка доступа к БД для всех файлов репозитория
var GlobalDB *sql.DB

// InitDB инициализирует подключение (вызывается из main.go)
func InitDB(db *sql.DB) {
	GlobalDB = db
}