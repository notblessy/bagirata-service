package model

type Dataset struct {
    ID    uint   `gorm:"primaryKey;autoIncrement"`
    Model string `gorm:"type:text"`
}