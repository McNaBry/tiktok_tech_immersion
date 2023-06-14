# Tiktok Assignment 2023

![Tests](https://github.com/TikTokTechImmersion/assignment_demo_2023/actions/workflows/test.yml/badge.svg)

The goal of the assignment is to modify the RPC server such that it can store and retrieve messages from a database, which for this assignment will be a local Redis instance created in the docker container.

To create and expose the Redis database instance on port 6379, the following code was added to the _docker-compose.yml_
```
redis:
    image: 'bitnami/redis:latest'
    environment:
      - ALLOW_EMPTY_PASSWORD=yes
    ports:
      - "6379:6379"
```

To handle interactions with the Redis instance, a new file (_redis_client.go_) was created in the rpc-server folder. To use Redis in Go, the module was imported from "github.com/redis/go-redis/v9". Documentation on the usage of Redis in Go can be found here: https://github.com/redis/go-redis. On startup, the RPC server will attempt to connect to the Redis instance on port 6379, via ```redis.NewClient(...)```.

SEND REQUEST
1. To process a send request, a unique identifier must first be assigned to the request so that it can be retrieved later from the Redis instance. 
2. This UID will be the chat attribute of the request (e.g. a1:a2) where it contains the sender and reciever. 
3. To extract and validate this UID, the function ```getRoomID(...)``` is used. 
4. As the roles of the sender and receiever can be reversed (e.g. a1:a2 -> a2:a1), the user with the lower alphanumeric value will be used as the first while the other will be the second part of the UID
5. Finally, the request (in the form of a message) will be added to the Redis instance using its UID. For each UID, there will be a list assigned to it which contains all the messages that has been sent.
6. To add the message to the instance, Redis' ZADD is used as this allows us to sort the messages by their timestamp which allows for easier retrieval later on

PULL REQUEST
1. To process a pull request, the UID must be extracted as well using the function ```getRoomID(...)```
2. The UID is then passed into ```RedisGetMessagesByRoomID(...)``` and other parameters - starting and ending index and whether to retrieve the messages in the reversed order
3. The messages are then sent back to the HTTP server which will feed the messages back to the user.
