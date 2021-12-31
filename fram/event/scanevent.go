package event

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"searchproxy/fram/utils"
	"strings"
)

type ScanConfig struct {
	ScanName string `json:"scanname"`
	CmdKey   string `json:"cmdkey"`
	Cmd      string `json:"cmd"`
	Out      string `json:"out"`
	Next     []Next `json:"next"`
}

type Next struct {
	Carry []string               `json:"carry"`
	Topic string                 `json:"topic"`
	Give  map[string]interface{} `json:"give"`
}

type ScanEvent struct {
	ecg *ScanConfig
}

func NewScanEvent(cfg *ScanConfig) (*ScanEvent, error) {
	sc := new(ScanEvent)
	sc.ecg = cfg
	return sc, nil
}

func (se ScanEvent) Action(data map[string]interface{}, pub Publish) error {
	_ = os.Remove(se.ecg.Out)
	cmd, err := se.cmdKeys(data)
	if err == nil {
		err := se.execCommand(cmd)
		if err != nil {
			_ = os.Remove(se.ecg.Out)
			return err
		} else {
			bts, err := ioutil.ReadFile(se.ecg.Out)
			_ = os.Remove(se.ecg.Out)
			if err != nil {
				fmt.Println("读取扫描结果出错", se.ecg.ScanName)
				return nil
			}
			if strings.TrimSpace(string(bts)) == "" {
				return nil
			}
			var btsObj interface{}
			utils.FatalAssert(json.Unmarshal(bts, &btsObj))
			if reflect.TypeOf(btsObj).Kind() == reflect.Slice {
				for _, b := range btsObj.([]interface{}) {
					tmp := b.(map[string]interface{})
					for _, v := range se.ecg.Next {
						// 携带参数
						for _, k := range v.Carry {
							if carryVul, ok := data[k]; ok {
								tmp[k] = carryVul
							}
						}
						tmp["scanname"] = se.ecg.ScanName
						// 额外参数
						for gk, gv := range v.Give {
							tmp[gk] = gv
						}
						nextMsg, err := json.Marshal(tmp)
						utils.FatalAssert(err)
						pub.PublishMsg(v.Topic, nextMsg)
					}
				}
			} else {
				for _, v := range se.ecg.Next {
					btsObj.(map[string]interface{})["scanname"] = se.ecg.ScanName
					for _, k := range v.Carry {
						if carryVul, ok := data[k]; ok {
							btsObj.(map[string]interface{})[k] = carryVul
						}
					}
					// 额外参数
					for gk, gv := range v.Give {
						btsObj.(map[string]interface{})[gk] = gv
					}
					_, _ = json.Marshal(btsObj)
					nextMsg, err := json.Marshal(btsObj)
					utils.FatalAssert(err)
					pub.PublishMsg(v.Topic, nextMsg)
				}
			}
		}
	} else {
		fmt.Println("获取命令行出错 ", se.ecg.ScanName, err)
		_ = os.Remove(se.ecg.Out)
		return err
	}
	_ = os.Remove(se.ecg.Out)
	return nil
}

func (se ScanEvent) cmdKeys(data map[string]interface{}) (string, error) {
	klist := strings.Split(strings.TrimSpace(se.ecg.CmdKey), " ")

	result := make(map[string]interface{})
	for _, v := range klist {
		if len(v) <= 0 || v == " " {
			continue
		}
		if vlu, ok := data[v]; ok {
			result[v] = vlu
		} else {
			return "", fmt.Errorf(fmt.Sprintf("key %s is nil", v))
		}
	}

	cmdStr := se.ecg.Cmd
	for k, v := range result {
		cmdStr = strings.Replace(cmdStr, fmt.Sprintf("{%v}", k), fmt.Sprintf("%v", v), -1)
	}
	return cmdStr, nil
}

func (se ScanEvent) execCommand(shell string) error {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	log.Println("开始执行 ", shell)
	cmd := exec.Command("/bin/sh", "-c", shell)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
