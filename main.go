package main

import (
	"searchproxy/fram/config"
	"searchproxy/fram/event"
	"searchproxy/fram/utils"
)

func main() {
	var mqcfg event.PublishConfig
	config.Install().Get("mq",&mqcfg)
	mqcfg.Topic=config.Install().GetScanName()
	var taskcfg event.TaskConfig
	config.Install().Get(config.Install().GetScanName(),&taskcfg)
	task, err := event.NewTask(&taskcfg)
	utils.FatalAssert(err)
	mq, err := event.NewPublish(&mqcfg,task)
	utils.FatalAssert(err)
	mq.Job()
}
