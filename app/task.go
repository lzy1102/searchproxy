package main

import (
	"searchproxy/app/fram/config"
	event2 "searchproxy/app/fram/event"
	"searchproxy/app/fram/utils"
)

func main() {
	var mqcfg event2.PublishConfig
	config.Install().Get("mq", &mqcfg)
	mqcfg.Topic = config.Install().GetScanName()
	var taskcfg event2.TaskConfig
	config.Install().Get(config.Install().GetScanName(), &taskcfg)
	task, err := event2.NewTask(&taskcfg)
	utils.FatalAssert(err)
	mq, err := event2.NewPublish(&mqcfg, task)
	utils.FatalAssert(err)
	mq.Job()
}
