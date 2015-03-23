// Copyright (c) 2015 by Doug Watson. MIT permissive licence - Contact info at http://github.com/dougwatson/http2amqp
package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var flagtests = []struct {
	queueName  string
	postData   string
	result     string
	statusCode int
}{
	{"AQUEUE", "Hello World", "SENT", 200},
	{"NOT_A_QUEUE", "Hello World", "BAD_QUEUE_NAME", 400},
}

func TestMessageDelivery(t *testing.T) {
	cli := &http.Client{}
	myWriter := bufio.NewWriter(logFile)
	lines := writeRabbit("amqp://guest:guest@localhost:5672/TEST", myWriter)

	//for testing, start a fake http server here on a random local port
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lines <- parseRequest(r)
		webReply(<-lines, w)
	}))
	defer ts.Close()

	for _, tt := range flagtests {
		req, reqErr := http.NewRequest("POST", ts.URL+"/"+tt.queueName, bytes.NewReader([]byte(tt.postData)))
		if reqErr != nil {
			t.Errorf("error=%v", reqErr)
		}
		resp, err := cli.Do(req)
		if err != nil {
			t.Errorf("error=%v", err)
		}

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		body := string(bodyBytes[:])
		if body != tt.result && resp.StatusCode == tt.statusCode {
			t.Errorf("result=[%s], want [%s] [%d]", body, tt.result, tt.statusCode)
		}
	}
}
