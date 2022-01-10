package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/imroc/req"
	"io/ioutil"
	"log"
	"os"
	"searchproxy/fram/utils"
	"sync"
)

type flagValue struct {
	mod     string
	scan    string
	cfgaddr string
	cfg     map[string]interface{}
}

var once sync.Once
var f *flagValue

func Install() *flagValue {
	once.Do(func() {
		f = &flagValue{}
		flag.StringVar(&f.mod, "mod", "0", "mod is release")
		flag.StringVar(&f.scan, "scan", "scanproxy", "config key name ")
		flag.StringVar(&f.cfgaddr, "cfgaddr", "config-1:8080", "config host:port")
		flag.Parse()
		var err error
		var out []byte
		if os.Getenv("MOD") != "" && os.Getenv("MOD") == "1" {
			f.mod = os.Getenv("MOD")
		}
		if os.Getenv("MOD") != "" && os.Getenv("MOD") == "0" {
			f.mod = os.Getenv("MOD")
		}
		if f.mod == "1" {
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

func (f *flagValue) Get(path string, obj interface{}) {
	log.Println(f.cfg)
	out, err := json.Marshal(f.cfg[path])
	utils.FatalAssert(err)
	_ = json.Unmarshal(out, obj)
	log.Println(obj)
}

func (f *flagValue) Mod() bool {
	return f.mod == "1"
}

func (f *flagValue) GetScanName() string {
	return f.scan
}
