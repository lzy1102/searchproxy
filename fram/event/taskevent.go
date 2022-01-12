package event

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"searchproxy/fram/logs"
	"searchproxy/fram/utils"
	"strings"
)

type TaskConfig struct {
	Datafile string   `json:"datafile"` // 传入的数据是否缓存文件
	TaskName string   `json:"taskname"`
	CmdKey   []string `json:"cmdkey"`
	Cmd      string   `json:"cmd"`
	Out      string   `json:"out"`
	Next     []Next   `json:"next"`
}

type Next struct {
	Carry []string               `json:"carry"` // 从上级携带给下级的参数
	Topic string                 `json:"topic"` // 投递给下级任务
	Give  map[string]interface{} `json:"give"`  // 额外携带给下级参数，如果必要的话，携带给下级
}

type TaskEvent struct {
	ecg *TaskConfig
}

func NewTask(cfg *TaskConfig) (*TaskEvent, error) {
	sc := new(TaskEvent)
	sc.ecg = cfg
	return sc, nil
}

func (t *TaskEvent) Action(data map[string]interface{}, pub Publish) error {
	if t.ecg.Out != "" {
		_ = os.Remove(t.ecg.Out)
		logs.Install().Infoln("删除输出文件", t.ecg.Out)
	}
	if t.ecg.Datafile != "" {
		databts, err := json.Marshal(data)
		if err != nil {
			return nil
		}
		err = ioutil.WriteFile(t.ecg.Datafile, databts, 0777)
		if err != nil {
			return nil
		}
		logs.Install().Infoln("写入数据源")
	}
	cmd, err := t.cmdKeys(data)
	if err == nil {
		err := t.execCommand(cmd)
		logs.Install().Infoln("执行命令结束")
		if err != nil {
			logs.Install().Errorln("执行命令结束,错误，重投")
			return err
		} else {
			bts, err := ioutil.ReadFile(t.ecg.Out)
			if err != nil {
				logs.Install().Errorln("扫描结果为空", t.ecg.TaskName)
				return nil
			}
			if strings.TrimSpace(string(bts)) == "" || strings.TrimSpace(string(bts)) == "null" {
				return nil
			}
			var btsObj interface{}
			utils.FatalAssert(json.Unmarshal(bts, &btsObj))
			logs.Install().Infoln("读取扫描结果", btsObj)
			if reflect.TypeOf(btsObj).Kind() == reflect.Slice {
				for _, b := range btsObj.([]interface{}) {
					tmp := b.(map[string]interface{})
					for _, v := range t.ecg.Next {
						// 携带参数
						for _, k := range v.Carry {
							if carryVul, ok := data[k]; ok {
								tmp[k] = carryVul
							}
						}
						tmp["taskname"] = t.ecg.TaskName
						// 额外参数
						for gk, gv := range v.Give {
							tmp[gk] = gv
						}
						nextMsg, err := json.Marshal(tmp)
						utils.FatalAssert(err)
						logs.Install().Infoln("投递给下级", v.Topic, "消息体", string(nextMsg))
						pub.PublishMsg(v.Topic, nextMsg)
					}
				}
			} else {
				for _, v := range t.ecg.Next {
					btsObj.(map[string]interface{})["taskname"] = t.ecg.TaskName
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
		logs.Install().Infoln("获取命令行出错 ", t.ecg.TaskName, err)
		return err
	}
	return nil
}

func (t *TaskEvent) cmdKeys(data map[string]interface{}) (string, error) {
	result := make(map[string]interface{})
	for _, v := range t.ecg.CmdKey {
		if len(strings.TrimSpace(v)) <= 0 || strings.TrimSpace(v) == " " {
			continue
		}
		if vlu, ok := data[v]; ok {
			result[v] = vlu
		} else {
			return "", fmt.Errorf(fmt.Sprintf("key %s is nil", v))
		}
	}

	cmdStr := t.ecg.Cmd
	for k, v := range result {
		cmdStr = strings.Replace(cmdStr, fmt.Sprintf("{%v}", k), fmt.Sprintf("%v", v), -1)
	}
	return cmdStr, nil
}

func (t *TaskEvent) execCommand(shell string) error {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序

	cmd := &exec.Cmd{}
	if runtime.GOOS == "linux" {
		cmd = exec.Command("sh", "-c", shell)
		logs.Install().Info("开始执行 ", shell)
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/c", shell)
		//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	piper, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	//开始执行命令
	err = cmd.Start()
	if err != nil {
		return err
	}
	//使用bufio包封装的方法实现对reader的读取
	reader := bufio.NewReader(piper)
	for {
		//换行分隔
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		//打印内容
		fmt.Println(line)
	}
	//等待命令结束并回收子进程资源等
	err = cmd.Wait()
	if err != nil {
		logs.Install().Errorln(err)
		return err
	}

	//err := cmd.Run()
	//if err != nil {
	//	return err
	//}
	return nil
}
