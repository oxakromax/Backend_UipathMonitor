package Structs

import "gorm.io/gorm"

type Handler struct {
	Db                  *gorm.DB
	TokenKey            string
	UniversalRoutes     []string
	UserUniversalRoutes []string
	DbKey               string
}
