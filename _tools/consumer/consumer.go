// Copyright (c) 2015 by Doug Watson. MIT permissive licence - Contact info at http://github.com/dougwatson/http2amqp
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/streadway/amqp"
)

var rabbitUri = flag.String("rabbitUri", "amqp://guest:guest@localhost:5672/", " or amqp://app:app@localhost:5672/")
var rabbitVHost = flag.String("rabbitVHost", "TEST", "Rabbitmq vhost, TEST or /")
var queueName = flag.String("queueName", "AQUEUE", "Queue name")
var block = flag.Bool("block", false, "use block, default true - YAGNI")

func main() {
	flag.Parse()
	if *block {
		readRabbitBlock()
	} else {
		readRabbitConsume()
	}
}
func readRabbitConsume() {
	amqpURI := *rabbitUri + *rabbitVHost
	conn, _ := amqp.Dial(amqpURI)
	channel, _ := conn.Channel()
	deliveries, _ := channel.Consume(
		*queueName, //queue name
		"tag",      //consumerTag
		true,       //autoAck
		false,      //exclusive
		false,      //nolocal
		false,      //noWait
		nil)        //arguments
	for d := range deliveries {
		fmt.Printf("got %s", d.Body)
	}
	//curl -H "content-type: application/json" -u guest:guest
	// -d '{"count": "1000","encoding": "auto", "queue": "AQUEUE", "vhost": "TEST", "arguments": {}, "requeue": "false"}'
	// http://localhost:15672/api/queues/%2F/FAKE_NOTIFY/get
}
func readRabbitBlock() {
	scheme := "http"
	host := "localhost:15672"
	path := "/api/queues/TEST/AQUEUE/get" //"/api/queues/%2F/AQUEUE/get"
	postData := fmt.Sprintf("{\"count\": \"10000\",\"encoding\": \"auto\", \"queue\": \"AQUEUE\", \"vhost\": \"%s\", \"arguments\": {}, \"requeue\": \"false\"}", *rabbitVHost)
	fmt.Printf("postData=%s", postData)
	req, _ := http.NewRequest("POST", "", bytes.NewReader([]byte(postData)))
	req.URL = &url.URL{
		Scheme: scheme,
		Host:   host,
		Opaque: path,
	}
	req.Header.Set("User-Agent", "http2amqp") //default
	req.SetBasicAuth("guest", "guest")        //user,pass
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		fmt.Printf("error=%v\n", err)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("body=%s", string(body[:10]))
	}
}
