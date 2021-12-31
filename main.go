package main

import (
	"searchproxy/fram/event"
	"searchproxy/fram/utils"
)

func main() {
	mq, err := event.NewMQ()
	utils.FatalAssert(err)
	mq.Job()
}
