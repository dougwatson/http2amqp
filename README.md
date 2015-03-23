http2amqp
===

```bash
$ ./http2amqp -h
Usage of ./http2amqp:
  -uri="amqp://guest:guest@localhost:5672/TEST": Address for the amqp or rabbitmq server (including vhost)
  ```

##RabbitMQ installation
*For Mac:
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
