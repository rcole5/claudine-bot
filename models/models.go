package models

type Command struct {
	Trigger string `gorm:"primary_key"`
	Action  string
	Channel string `gorm:"primary_key"`
}
