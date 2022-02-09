package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/imroc/req"
	"io/ioutil"
	"os"
	"searchproxy/app/fram/utils"
	"sync"
)

type flagValue struct {
	mod     bool
	scan    string
	cfgaddr string
	cfg     map[string]interface{}
}

var once sync.Once
var f *flagValue

func Install() *flagValue {
	once.Do(func() {
		f = &flagValue{}
		flag.BoolVar(&f.mod, "mod", false, "mod is release")
		flag.StringVar(&f.scan, "scan", "scanproxy", "config key name ")
		flag.StringVar(&f.cfgaddr, "cfgaddr", "config-1:8080", "config host:port")
		flag.Parse()
		var err error
		var out []byte
		if os.Getenv("MOD") != "" && os.Getenv("MOD") == "1" {
			f.mod = true
		}
		if os.Getenv("MOD") != "" && os.Getenv("MOD") == "0" {
			f.mod = false
		}
		if f.mod {
			//out, err = ioutil.ReadFile(fmt.Sprintf("%v/config.json",utils.GetCurrentAbPathByExecutable()))
			r, err := req.Get(fmt.Sprintf("http://%v/api/config/get", f.cfgaddr))
			if err != nil {
				return
			}
			out = r.Bytes()
		} else {
			out, err = ioutil.ReadFile("config.json")
		}
		utils.FatalAssert(err)
		utils.FatalAssert(json.Unmarshal(out, &f.cfg))
	})
	return f
}

func (f flagValue) Reget(path string, obj interface{}) {
	r, err := req.Get(fmt.Sprintf("http://%v/api/config/get", f.cfgaddr))
	utils.FatalAssert(err)
	utils.FatalAssert(json.Unmarshal(r.Bytes(), &f.cfg))
	out, err := json.Marshal(f.cfg[path])
	utils.FatalAssert(err)
	_ = json.Unmarshal(out, obj)
}

func (f flagValue) Get(path string, obj interface{}) {
	out, err := json.Marshal(f.cfg[path])
	utils.FatalAssert(err)
	_ = json.Unmarshal(out, obj)
}

func (f flagValue) Mod() bool {
	return f.mod
}

func (f flagValue) GetScanName() string {
	return f.scan
}
