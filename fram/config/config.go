package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"searchproxy/fram/utils"
	"sync"
)

type flagValue struct {
	mod bool
	scan string
	cfg map[string]interface{}
}

var once sync.Once
var f *flagValue

func Install() *flagValue {
	once.Do(func() {
		f = &flagValue{}
		flag.BoolVar(&f.mod, "mod", false, "mod is release")
		flag.StringVar(&f.scan, "scan", "scanproxy", "config key name ")
		flag.Parse()
		var err error
		var out  []byte
		if f.mod {
			out, err = ioutil.ReadFile(fmt.Sprintf("%v/config.json",utils.GetCurrentAbPathByExecutable()))
		}else {
			out, err = ioutil.ReadFile("config.json")
		}
		utils.FatalAssert(err)
		utils.FatalAssert(json.Unmarshal(out, &f.cfg))
	})
	return f
}

func (f *flagValue) Get(path string,obj interface{}) {
	out, err := json.Marshal(f.cfg[path])
	utils.FatalAssert(err)
	_ = json.Unmarshal(out, obj)
}

func (f *flagValue) Mod() bool {
	return f.mod
}

func (f *flagValue) GetScanName() string  {
	return f.scan
}