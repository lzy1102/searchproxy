package info

import "gorm.io/gorm"

type Data struct {
	gorm.Model
	Ip        string `json:"ip"`
	Port      string `json:"port"`
}