package main

import (
	"encoding/json"
	"flag"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"searchproxy/fram/config"
	db2 "searchproxy/fram/db"
	"searchproxy/fram/utils"
)

type save struct {
	datafile string
	db       *db2.Models
	cache    *db2.RedisClient
}

var sc save

func init() {
	flag.StringVar(&sc.datafile, "file", "datafile.json", "")
	flag.Parse()
}


func main() {
	var dbcfg db2.MongoConfig
	config.Install().Get("mongo", &dbcfg)
	sc.db = db2.NewMongo(&dbcfg)

	var cachecfg db2.RedisConfig
	config.Install().Get("redis", &cachecfg)
	sc.cache = db2.NewRedis(&cachecfg)

	bts, err := ioutil.ReadFile(sc.datafile)
	utils.FatalAssert(err)
	var data map[string]interface{}
	utils.FatalAssert(json.Unmarshal(bts, &data))

	if ipstr,ok:= data["ip"];ok {
		data["ip"]=utils.Ip2Int64(ipstr.(string))
	}
	var result map[string]interface{}
	_ = sc.db.FindOne("info", bson.M{"ip": data["ip"], "port": data["port"]}, &result)
	
}
