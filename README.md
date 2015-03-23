http2amqp
===

```bash
$ ./http2amqp -h
Usage of ./http2amqp:
  -httpPort="8080": The listen port for the https GET requests
  -uri="amqp://guest:guest@localhost:5672/TEST": Address for the amqp or rabbitmq server (including vhost)
  ```
## Goals
Allows a stateless service (like a php page) to make fast queue inserts into an AMQP queue without the overhead of re-establishing the connection.

Supports auto-reconnect and re-synchronization of client and server.

##RabbitMQ installation
For Mac:
```bash
brew install rabbitmq
```
Other OS, install from source: https://www.rabbitmq.com/releases/rabbitmq-server/v3.4.4/rabbitmq-server-mac-standalone-3.4.4.tar.gz

##Startup rabbitmq:
```bash
nohup rabbitmq-server &
```

##Define your message queue:

```bash
rabbitmqctl add_vhost TEST
rabbitmqctl set_permissions -p TEST guest ".*" ".*" ".*"
rabbitmqadmin declare queue --vhost=TEST name=AQUEUE durable=true
```

##Check your new message queue:
```bash
rabbitmqctl  list_vhosts
rabbitmqctl  list_permissions -p TEST
rabbitmqctl  list_bindings -p TEST
rabbitmqctl  list_queues -p TEST
```
##To test:
```bash
nohup ./http2amqp &
curl localhost:8080/AQUEUE/hello1
curl localhost:8080/AQUEUE/hello2
curl localhost:8080/AQUEUE/hello3
```

##Verify:
```bash
_tools/consumer/consumer
got Hello1
got hello2
got hello3
```
