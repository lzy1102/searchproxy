package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"searchproxy/fram/utils"
	"sync"
)

type flagValue struct {
	mod bool
	cfg map[string]interface{}
}

var once sync.Once
var f *flagValue

func Install() *flagValue {
	once.Do(func() {
		f = &flagValue{mod:false}
		var err error
		var out  []byte
		if f.mod {
			out, err = ioutil.ReadFile(fmt.Sprintf("%v/config.json",utils.GetCurrentAbPathByExecutable()))
		}else {
			out, err = ioutil.ReadFile("config.json")
		}
		utils.FatalAssert(err)
		utils.FatalAssert(json.Unmarshal(out, f.cfg))
	})
	return f
}

func (f *flagValue) Get(path string,obj interface{}) {

}