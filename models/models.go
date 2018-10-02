package models

import "github.com/jinzhu/gorm"

type Command struct {
	gorm.Model
	Trigger string
	Action  string
	Channel string
}
