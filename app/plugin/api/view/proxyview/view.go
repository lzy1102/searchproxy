package proxyview

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"searchproxy/app/fram/db"
)

func GetProxylist(ctx *gin.Context) {
	var data []interface{}
	db.MongoInstance().FindManyLimit("info", bson.M{}, &data, 10, 0)
	ctx.JSON(200, data)
}
