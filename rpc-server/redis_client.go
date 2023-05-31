//https://tutorialedge.net/golang/go-redis-tutorial/

package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

// Holds the RedisClient connection
type RedisClient struct {
	client *redis.Client
}

// Defines the template for a chat message
type Message struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// Initalizes the connection to the Redis db
// And assigns it to the RedisClient passed in
func (c *RedisClient) ConnectToRedis(ctx context.Context, address string, password string) error {
	log.Println("Creating Redis connection...")

	// Create the redis connection
	temp := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	// Ping the connection and get the result
	pong, err := temp.Ping(ctx).Result()
	// Print the result
	log.Println(pong, err)
	if err == redis.Nil {
		return err
	}

	// Assign connection to the RedisClient passed in
	c.client = temp
	return nil
}

func (c *RedisClient) RedisSaveMessage(ctx context.Context, roomID string, message *Message) error {
	log.Printf("Saving message to %v", roomID)
	// Get JSON representation of message
	text, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Member to be added to the list
	member := &redis.Z{
		Score:  message.Timestamp, // Sort key for ZADD
		Member: text,
	}

	// Redis ZADD, add chat message to list with the corresponding roomID
	_, err = c.client.ZAdd(ctx, roomID, *member).Result()
	if err != nil {
		return err
	}

	return nil
}

// Retrieve array of Messages belonging to a roomID
func (c *RedisClient) RedisGetMessagesByRoomID(ctx context.Context, roomID string, start, end int64, reverse bool) ([]*Message, error) {
	var (
		tempMessages []string
		messages     []*Message
		err          error
	)

	// Returns either in ascending or descending order
	if reverse {
		tempMessages, err = c.client.ZRevRange(ctx, roomID, start, end).Result()
		if err != nil {
			return nil, err
		}
	} else {
		tempMessages, err = c.client.ZRange(ctx, roomID, start, end).Result()
		if err != nil {
			return nil, err
		}
	}

	for _, msg := range tempMessages {
		temp := &Message{}
		// Reverses JSON encoding and
		// Transforms the JSON data into the form given by temp
		// Which in this case is the Message struct
		err := json.Unmarshal([]byte(msg), temp)
		if err != nil {
			return nil, err
		}
		// Append transformed message into message array
		messages = append(messages, temp)
	}

	return messages, nil
}
