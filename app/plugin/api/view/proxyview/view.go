package proxyview

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"searchproxy/app/fram/config"
	"searchproxy/app/fram/db"
	"searchproxy/app/fram/utils"
	"strconv"
)

func GetProxylist(ctx *gin.Context) {
	google := ctx.Query("google")
	protocol := ctx.Query("protocol")
	limitstr := ctx.Query("limit")
	skipstr := ctx.Query("skip")
	filter := bson.M{}
	if google != "" {
		if google == "1" {
			filter["google"] = true
		} else if google == "0" {
			filter["google"] = false
		}
	}
	if protocol != "" {
		filter["protocol"] = protocol
	}
	limit, err := strconv.ParseInt(limitstr, 10, 64)
	skip, err := strconv.ParseInt(skipstr, 10, 64)
	var data []interface{}
	err = db.MongoInstance().FindManyLimit("info", filter, &data, limit, skip)
	if err != nil {
		return
	}
	for _, ele := range data {
		ele.(map[string]interface{})["ip"] = utils.Int64ToIp(ele.(map[string]interface{})["ip"].(int64))
	}
	ctx.JSON(200, data)
}

func DeleteProxy(ctx *gin.Context) {
	id := ctx.PostForm("id")
	hex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return
	}
	db.MongoInstance().DeleteMany("info", bson.M{"_id": hex})
	ctx.JSON(200, gin.H{"code": 200})
}

func AddPort(ctx *gin.Context) {
	port := ctx.PostForm("port")
	var cfg map[string]interface{}
	json.Unmarshal(config.Install().RegetAll(), &cfg)

	if ports, ok := cfg["ports"]; ok {
		have := false
		for _, i := range ports.([]interface{}) {
			if fmt.Sprintf("%v", i) == port {
				have = true
			}
		}
		if have == false {
			ports = append(ports.([]interface{}), port)
		}
	}
	config.Install().Reset(cfg)
	ctx.JSON(200, gin.H{"code": 200, "data": cfg})
}
