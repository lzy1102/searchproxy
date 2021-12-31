package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/imroc/req"
	"io/ioutil"
	"log"
	"searchproxy/fram/config"
	"searchproxy/fram/info"
	"searchproxy/fram/utils"
	"sync"
)

type fofacfg struct {
	email string
	key   string
}

var cfg *fofacfg
var on sync.Once
var store map[string]bool

func Install() *fofacfg {
	on.Do(func() {
		cfg = &fofacfg{}
		cfg.email = config.Install().Get("fofa").(map[string]interface{})["email"].(string)
		cfg.key = config.Install().Get("fofa").(map[string]interface{})["token"].(string)
	})
	return cfg
}


func main() {
	c:=fofacfg{email:"yeshenmei@163.com",key:"c9581ffaeb3e8320e54633394c3a4d19"}
	rst:= c.FoFa(`protocol=="socks5" && "Version:5 Method:No Authentication(0x00)" && after="2021-08-01"`)
	bts, _ := json.Marshal(rst)
	_ = ioutil.WriteFile("fofa.json", bts, 0777)
	for i, i2 := range rst {
		log.Println(i,i2.Ip,i2.Port)
	}
}

func (c *fofacfg) FoFa(ctx string) []info.Data {
	store = make(map[string]bool)
	c.registered()
	searchStr := base64.StdEncoding.EncodeToString([]byte(ctx))
	page := 1
	var addrlist []info.Data
	for {
		tmp := c.search(searchStr, page)
		if len(tmp) == 0 {
			break
		}
		addrlist = append(addrlist, tmp...)
		page++
		if page >=1{
			break
		}
	}
	return addrlist
}


func (c *fofacfg) registered() {
	params := req.QueryParam{
		`email`: c.email,
		`key`:   c.key,
	}
	response, _ := req.Get(`https://fofa.so/api/v1/info/my`, params)
	var data map[string]interface{}
	utils.FatalAssert(response.ToJSON(&data))
	c.errlog(data)
}

func (c *fofacfg) search(sh string, page int) []info.Data {
	params := req.QueryParam{
		`email`:   c.email,
		`key`:     c.key,
		`qbase64`: sh,
		`page`:    page,
		`size`:    99,
	}
	response, _ := req.Get(`https://fofa.so/api/v1/search/all`, params)
	var data map[string]interface{}
	utils.FatalAssert(response.ToJSON(&data))
	if c.errlog(data) {
		return []info.Data{}
	}
	results := data["results"].([]interface{})
	var addrlist []info.Data
	for _, i2 := range results {
		addr := i2.([]interface{})
		if _, ok := store[fmt.Sprintf("%v:%v", addr[1], addr[2])]; !ok {
			store[fmt.Sprintf("%v:%v", addr[1], addr[2])] = true
			addrlist = append(addrlist, info.Data{
				Ip:      fmt.Sprintf("%v",addr[1]) ,
				Port:    fmt.Sprintf("%v",addr[2]),
			})
		}
	}
	return addrlist
}

func (c *fofacfg) errlog(data map[string]interface{}) bool {
	if errorinfo, ok := data["error"]; ok {
		return errorinfo.(bool)
	}
	return false
}
