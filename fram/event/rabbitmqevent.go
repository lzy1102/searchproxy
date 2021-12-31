package event

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"searchproxy/fram/utils"
)

type MQConfig struct {
	Topic string `json:"topic"`
	Uri   string `json:"uri"`
}

type RabbitMqEvent struct {
	config   *MQConfig
	cconn    *amqp.Connection
	consumer *amqp.Channel

	pconn    *amqp.Connection
	producer *amqp.Channel
	task     Task
}

func NewMQ(cfg *MQConfig, task Task) (*RabbitMqEvent, error) {
	single := new(RabbitMqEvent)
	single.config = cfg
	single.task = task
	return single, nil
}

func (r RabbitMqEvent) closeProducer() {
	if r.producer != nil {
		utils.FatalAssert(r.producer.Close())
	}
	if r.pconn != nil {
		utils.FatalAssert(r.pconn.Close())
	}
}

func (r RabbitMqEvent) newProducer() {
	r.closeProducer()
	var err error
	r.pconn, err = amqp.Dial(r.config.Uri)
	utils.FatalAssert(err)
	r.producer, err = r.pconn.Channel()
	utils.FatalAssert(err)
}

func (r RabbitMqEvent) closeConsumer() {
	if r.consumer != nil {
		utils.FatalAssert(r.consumer.Close())
	}
	if r.cconn != nil {
		utils.FatalAssert(r.cconn.Close())
	}
}

func (r RabbitMqEvent) newConsumer() {
	r.closeConsumer()
	var err error
	r.cconn, err = amqp.Dial(r.config.Uri)
	utils.FatalAssert(err)
	r.consumer, err = r.cconn.Channel()
	utils.FatalAssert(err)
}

func (r RabbitMqEvent) registerConsumer() (<-chan amqp.Delivery, error) {
	que, err := r.consumer.QueueDeclare(r.config.Topic, true, false, false,
		false, nil)
	if nil != err {
		return nil, fmt.Errorf("failed to queue declare, cuz %s", err.Error())
	}
	if err := r.consumer.Qos(1, 0, false); nil != err {
		return nil, fmt.Errorf("failed to set qos, cuz %s", err.Error())
	}
	msgs, err := r.consumer.Consume(que.Name, "", false,
		false, false, false, nil)
	if nil != err {
		return nil, fmt.Errorf("failed to register consumer, cuz %s", err.Error())
	}
	return msgs, nil
}

func (r RabbitMqEvent) PublishMsg(topic string, msg []byte) {
	q, err := r.producer.QueueDeclare(
		topic, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	utils.FatalAssert(err)
	err = r.producer.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         msg,
		})
	utils.FatalAssert(err)
}

func (r RabbitMqEvent) Job() {
	r.newProducer()
	defer r.closeProducer()
	r.newConsumer()
	defer r.closeConsumer()
	msgs, err := r.registerConsumer()
	utils.FatalAssert(err)
	for msg := range msgs {
		var data map[string]interface{}
		err=json.Unmarshal(msg.Body, &data)
		if err != nil {
			log.Println("消息转字典出错,抛弃")
			_ = msg.Ack(true)
			continue
		}
		log.Println("开始执行消息", r.config.Topic)
		err = r.task.Action(data, r)
		if err != nil {
			log.Println("执行消息出错，重投", r.config.Topic)
			_ = msg.Nack(false, true)
		}
		log.Println("消息执行结束", r.config.Topic)
		_ = msg.Ack(true)
	}
}
