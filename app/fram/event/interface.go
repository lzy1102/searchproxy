package event

type Publish interface {
	PublishMsg(string, []byte)
}

type Task interface {
	Action(map[string]interface{}, Publish) error
}
