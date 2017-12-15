// Copyright (c) 2015 by Doug Watson. MIT permissive licence - Contact info at http://github.com/dougwatson/http2amqp
package main

//usage:
// POST http://localhost:8080/QUEUE_NAME
// put the message in the body
import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"log"
)

var uri, httpPort, logFilePath string
var logFile *os.File
var queueMap = make(map[string]string)
var mutex = &sync.Mutex{}

func init() {
	flag.StringVar(&uri, "uri", "amqp://guest:guest@localhost:5672/TEST", "The address for the amqp server (including vhost)")
	flag.StringVar(&httpPort, "httpPort", "8080", "The listen port for the https GET requests")
	flag.StringVar(&logFilePath, "logFilePath", "/var/log/http2amqp.log", "Set path to get log information");
}

func main() {
	flag.Parse()
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Log error: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	myWriter := bufio.NewWriter(logFile)
	fmt.Fprintf(myWriter, "%v startup uri=%s\n", time.Now(), uri)
	log.Printf("%v startup uri=%s\n", time.Now(), uri)
	myWriter.Flush()
	lines := writeRabbit(uri, myWriter) //read device requests rabbitmq o
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		staticHandler(w, r, lines)
	})
	http.ListenAndServe(":"+httpPort, nil) //address= ":8080"
}

func staticHandler(w http.ResponseWriter, req *http.Request, lines chan string) {
	result := ""
	select {
	case lines <- parseRequest(req):
	case <-time.After(time.Second):
		result = "NETWORK_SEND_TIMEOUT|503"
		webReply(result, w)
		return
	}
	select {
	case result = <-lines:
	case <-time.After(time.Second):
		result = "NETWORK_REC_TIMEOUT|504"
	}
	webReply(result, w)
}
func webReply(result string, w http.ResponseWriter) {
	var statusMessage string
	var statusCode int
	resultArr := strings.Split(result, "|") //split into 2 parts- status message and status code
	if len(resultArr) == 2 {
		statusMessage = resultArr[0]
		statusCode, _ = strconv.Atoi(resultArr[1])
	} else {
		statusMessage, statusCode = "PARSE_ERROR", 510
	}
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, statusMessage)
	log.Printf(statusMessage, w)
}
func parseRequest(req *http.Request) string {
	urlPath := req.URL.Path[len("/"):] //take everything after the http://localhost:8080/ (it gets the queue name)
	bodyBytes, _ := ioutil.ReadAll(req.Body)
	body := string(bodyBytes[:])
	return urlPath + "/" + body //write to rabbitMQ
}

func writeRabbit(amqpURI string, myWriter *bufio.Writer) chan string {
	lines := make(chan string)
	go func() {
		connectionAttempts := 0
		for {
			conn, err1 := amqp.Dial(amqpURI)
			if err1 != nil {
				fmt.Printf("%v err1=%v\n", time.Now(), err1)
				log.Printf("%v err1=%v\n", time.Now(), err1)
				fmt.Fprintf(myWriter, "%v err1=%v\n", time.Now(), err1)
				log.Printf("%v err1=%v\n", time.Now(), err1)
				time.Sleep(time.Second)
				myWriter.Flush()
				continue
			}
			channel, err2 := conn.Channel()
			if err2 != nil {
				fmt.Fprintf(myWriter, "%v err2=%v\n", time.Now(), err2)
				log.Printf("%v err2=%v\n", time.Now(), err2)
			}
			i := 0
			go func() {
				fmt.Fprintf(myWriter, "%v %d %d closing (will reopen): %s\n", time.Now(), connectionAttempts, i, <-conn.NotifyClose(make(chan *amqp.Error)))
				log.Printf("%v %d %d closing (will reopen): %s\n", time.Now(), connectionAttempts, i, <-conn.NotifyClose(make(chan *amqp.Error)))
			}()

			myWriter.Flush()
			result := ""
			for {
				i++
				line := <-lines

				startTime := time.Now()

				urlPath := strings.SplitN(line, "/", 2) //split into 2 parts- queueName and Message
				if len(urlPath) < 2 {
					fmt.Fprintf(myWriter, "%v %d %d Skip this message b/c it is missing a QUEUE name on the URL or a message body. count=%d line=%v\n", time.Now(), connectionAttempts, i, len(urlPath), line)
					log.Printf("%v %d %d Skip this message b/c it is missing a QUEUE name on the URL or a message body. count=%d line=%v\n", time.Now(), connectionAttempts, i, len(urlPath), line)
					myWriter.Flush()
					lines <- "skip"
					continue
				}
				queue, message := urlPath[0], urlPath[1]

				if queueMap[queue] == "" {
					//if we have never seen this queue name work before
					_, err := channel.QueueInspect(queue)
					if err == nil {
						mutex.Lock()
						queueMap[queue] = "1" //save the successful lookup of the queue name
						mutex.Unlock()
					} else {
						result = "BAD_QUEUE_NAME|400"
						lines <- result
						fmt.Fprintf(myWriter, "%v %d %d %s/%db %s %.6f\n", time.Now(), connectionAttempts, i, queue, len(message), result, 1.0)
						log.Printf("%v %d %d %s/%db %s %.6f\n", time.Now(), connectionAttempts, i, queue, len(message), result, 1.0)
						break
					}

				}

				err3 := channel.Publish(
					"",    //exchange
					queue, //routingKey, for some reason we need to put the queue name here
					false, //mandatory - don't quietly drop messages in case of missing Queue
					false, //immediate
					amqp.Publishing{
						Headers:         amqp.Table{},
						ContentType:     "text/plain",
						ContentEncoding: "UTF-8",
						Body:            []byte(message),
						DeliveryMode:    amqp.Persistent, // 1=non-persistent(Transient), 2=persistent
						Priority:        9,
					},
				)
				if err3 != nil {
					result = "NETWORK_ERROR|502"
					lines <- result
					fmt.Fprintf(logFile, "%v %d %d \nerr3 saw a network error=%v\n", time.Now(), connectionAttempts, i, err3)
					log.Printf("%v %d %d \nerr3 saw a network error=%v\n", time.Now(), connectionAttempts, i, err3)
					myWriter.Flush()
					//TODO - we should put the message back onto the channel since the publish failed
					break //probably the connection broke due to a network issue, so break out of this loop so it will re-connect
				}

				result = "SENT|200"
				lines <- result
				duration := (time.Since(startTime)).Seconds()
				fmt.Fprintf(myWriter, "%v %d %d %s/%db %s %.6f\n", time.Now(), connectionAttempts, i, queue, len(message), result, duration)
				log.Printf("%v %d %d %s/%db %s %.6f\n", time.Now(), connectionAttempts, i, queue, len(message), result, duration)
				if result != "SENT|200" {
					break
				}

				myWriter.Flush()
			}

		}
	}()
	return lines
}
