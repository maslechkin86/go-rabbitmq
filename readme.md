# ⚙️ Install and setup RabbitMQ
This command will download the latest RabbitMQ and start it as a background process, exposing ports `5672` and `15672`.
````shell
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.11-management
````

### Create a new user
````shell
docker exec rabbitmq rabbitmqctl add_user user_name user_password
````

### Set a user admin permissions
````shell
docker exec rabbitmq rabbitmqctl set_user_tags user_name administrator
````

### Delete a user
```shell
docker exec rabbitmq rabbitmqctl delete_user user_name
```

### Create a new vhost
Creat a new vhost `customers` and set full permissions for `guest` user to the vhost. 

````shell
docker exec rabbitmq rabbitmqctl add_vhost customers
docker exec rabbitmq rabbitmqctl set_permissions -p customers guest ".*" ".*" ".*"
````

### Create a new exchange
Create a new exchange `customer_events` and set permissions for `guest` user.
````shell
docker exec rabbitmq rabbitmqadmin declare exchange --vhost=customers name=customer_events type=topic -u guest -p guest durable=true
docker exec rabbitmq rabbitmqctl set_topic_permissions -p customers guest customer_events "^customers.*" "^customers.*"
````

# 1️⃣ Publishing and Consuming Messages
![01.webp](resource%2F01.webp)
We are using FIFO Queues(First in First out). This means each message is only sent to one Consumer.
````shell
go run cmd/consumer01/main.go
go run cmd/producer01/main.go
````

# 2️⃣ Publish and Subscribe (PubSub)
![02.webp](resource%2F02.webp)
In a publish and subscribe schema, each consumer receives the same message.

Recreate an exchange `customer_events` with `Fanout` as the type and give full access for `guest` user.
````shell
docker exec rabbitmq rabbitmqadmin delete exchange name=customer_events --vhost=customers -u guest -p guest
docker exec rabbitmq rabbitmqadmin declare exchange --vhost=customers name=customer_events type=fanout -u guest -p guest durable=true
docker exec rabbitmq rabbitmqctl set_topic_permissions -p customers guest customer_events ".*" ".*"
````

Run two consumers app and one producer.
````shell
go run cmd/consumer02/main.go
go run cmd/producer02/main.go
````

# Remote procedure call (RPC)
![03.webp](resource%2F03.webp)
The producer receives an acknowledgement of the processing of a message from the consumer within a new message.

Creating a new exchange `customer_callbacks` with `Direct` type and adding full permissions for `guest` user.
```shell
docker exec rabbitmq rabbitmqadmin declare exchange --vhost=customers name=customer_callbacks type=direct -u guest -p guest durable=true
docker exec rabbitmq rabbitmqctl set_topic_permissions -p customers guest customer_callbacks ".*" ".*"
```