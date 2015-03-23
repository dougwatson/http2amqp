http2amqp
===

```bash
$ ./http2amqp -h
Usage of ./http2amqp:
  -uri="amqp://guest:guest@localhost:5672/TEST": The address for the amqp or rabbitmq server (including vhost)
  ```

#Installation
For Mac: brew install rabbitmq
Other OS: https://www.rabbitmq.com/releases/rabbitmq-server/v3.4.4/rabbitmq-server-mac-standalone-3.4.4.tar.gz

#Startup rabbitmq:
nohup rabbitmq-server &

#Define your message Queue:

rabbitmqctl add_vhost TEST
rabbitmqctl  set_permissions -p TEST guest ".*" ".*" ".*"
rabbitmqadmin declare queue --vhost=TEST name=AQUEUE durable=true

#Check your new message queue:
rabbitmqctl  list_vhosts
rabbitmqctl  list_permissions -p TEST
rabbitmqctl  list_bindings -p TEST
rabbitmqctl  list_queues -p TEST

