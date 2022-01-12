package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func main() {
	readFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		return
	}
	var data map[string]interface{}
	json.Unmarshal(readFile, &data)
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	app.GET("api/config/get", func(c *gin.Context) {
		c.JSON(200, data)
	})
	app.POST("api/config/set", func(c *gin.Context) {
		err := c.BindJSON(&data)
		if err != nil {
			return
		}
		marshal, err := json.Marshal(data)
		if err != nil {
			return
		}
		ioutil.WriteFile("config.json", marshal, 0777)
	})
	err = app.Run(":8080")
	if err != nil {
		return
	}
}
