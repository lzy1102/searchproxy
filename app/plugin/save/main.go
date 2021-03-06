package main

import (
	"encoding/json"
	"flag"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"os"
	"searchproxy/app/fram/config"
	"searchproxy/app/fram/db"
	"searchproxy/app/fram/utils"
	"time"
)

type save struct {
	datafile string
	db       *db.Models
	cache    *db.RedisClient
}

//func init() {
//flag.StringVar(&sc.datafile, "file", "datafile.json", "")
//flag.Parse()
//}

func main() {
	sc := save{}
	//sc.datafile = os.Args[len(os.Args)-1]
	sc.datafile = *flag.String("datafile", "datafile.json", "")

	bts, err := ioutil.ReadFile(sc.datafile)
	utils.FatalAssert(err)
	var data map[string]interface{}
	utils.FatalAssert(json.Unmarshal(bts, &data))

	var dbcfg db.MongoConfig
	config.Install().Get("mongo", &dbcfg)
	sc.db = db.NewMongo(&dbcfg)

	var cachecfg db.RedisConfig
	config.Install().Get("cache", &cachecfg)
	sc.cache = db.NewRedis(&cachecfg)
	if ipstr, ok := data["ip"]; ok {
		data["ip"] = utils.Ip2Int64(ipstr.(string))
	}
	var result map[string]interface{}
	_ = sc.db.FindOne("info", bson.M{"ip": data["ip"], "port": data["port"]}, &result)
	if id, ok := result["_id"]; ok {
		_, _ = sc.db.UpdateOne("info", bson.M{"_id": id}, data)
	} else {
		_, err = sc.db.InsertOne("info", data)
		if err == nil {
			get := sc.cache.GetInt64("proxycount")
			sc.cache.Set("proxycount", get+1, time.Hour*1000000)
		}
	}
	os.RemoveAll(sc.datafile)
	sc.db.Close()
	sc.cache.Close()
}
