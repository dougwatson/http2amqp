// Copyright (c) 2015 by Doug Watson. MIT permissive licence - Contact info at http://github.com/dougwatson/http2amqp
package main

import (
	"flag"
	"fmt"

	"github.com/streadway/amqp"
)

var count int
var uri string
var queue string

func init() {
	flag.IntVar(&count, "c", 10, "Number of messages to add")
	flag.StringVar(&uri, "uri", "amqp://guest:guest@localhost:5672/TEST", "The address for the queue (including vhost)")
	flag.StringVar(&queue, "q", "AQUEUE", "The queue")
}

func main() {
	flag.Parse()
	amqpURI := uri //"amqp://guest:guest@localhost:5672/TEST"
	conn, _ := amqp.Dial(amqpURI)
	channel, _ := conn.Channel()
	tenant := "toggle"
	for i := 0; i < count; i++ {
		//fmt.Printf("i=%s", "x"+fmt.Sprintf("%d", i))
		err := publish(channel, i, tenant, queue)
		if err == nil {
			fmt.Printf("OK published %d,%s\n", i, tenant)
		} else {
			fmt.Printf("ERR published %d,%s\n", i, tenant)
		}
	}
}
func publish(channel *amqp.Channel, i int, tenant string, queue string) (err error) {
	err = channel.Publish(
		"",    //exchange
		queue, //routingKey, for some reason we need to put the queue name here
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "UTF-8",
			Body:            []byte(fmt.Sprintf("%d,%s", i, tenant)),
			DeliveryMode:    2, // 1=non-persistent, 2=persistent
			Priority:        9,
		},
	)
	return
}
