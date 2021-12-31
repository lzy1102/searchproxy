package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"searchapi/fram/info"
	"searchapi/fram/utils"
	"sync"
)

type dbcfg struct {
	DB *gorm.DB
}

var cfg *dbcfg
var on sync.Once

func Install() *dbcfg {
	on.Do(func() {
		cfg = new(dbcfg)
		utils.FatalAssert(os.RemoveAll("db.db"))
		var err error
		cfg.DB, err = gorm.Open(sqlite.Open("db.db"), &gorm.Config{})
		utils.FatalAssert(err)
		utils.FatalAssert(cfg.DB.AutoMigrate(&info.Data{}))
	})
	return cfg
}

func main() {
	Install().DB.Create(&info.Data{
		Ip:        "127.0.0.1",
		Port:      "80",
		Title:     "",
		Resheader: "",
		Url:       "",
		Domain:    "",
	})
	Install().DB.Create(&info.Data{
		Ip:        "127.0.0.1",
		Port:      "443",
		Title:     "",
		Resheader: "",
		Url:       "",
		Domain:    "",
	})
	Install().DB.Create(&info.Data{
		Ip:        "1.1.1.1",
		Port:      "80",
		Title:     "",
		Resheader: "",
		Url:       "",
		Domain:    "",
	})
	var data info.Data
	Install().DB.Where(&info.Data{Ip: "127.0.0.1", Port: "80"}).Find(&data)
	log.Println(data.Ip, data.Port, data.Url)
	Install().DB.Model(&info.Data{}).Where(&info.Data{Ip: "127.0.0.1", Port: "80"}).Updates(info.Data{
		Url:       "http://127.0.0.1:80",
	})
	Install().DB.Where(&info.Data{Ip: "127.0.0.1", Port: "80"}).Find(&data)
	log.Println(data.Ip, data.Port, data.Url)
}
