package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	log.Println("rpc SendRequest")

	// Validate send request
	if err := valSendReq(req); err != nil {
		return nil, err
	}

	// Create timestamp, default - time this function is triggered
	timestamp := time.Now().Unix()
	message := &Message{
		Message:   req.Message.GetText(),
		Sender:    req.Message.GetSender(),
		Timestamp: timestamp,
	}

	roomID, err := getRoomID(req.Message.GetChat())
	if err != nil {
		return nil, err
	}

	err2 := RedisDB.RedisSaveMessage(ctx, roomID, message)
	if err2 != nil {
		return nil, err2
	}

	resp := rpc.NewSendResponse()
	resp.Code, resp.Msg = 0, "success"
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	log.Println("rpc PullRequest")
	// retrieve roomID
	roomID, err := getRoomID(req.GetChat())
	if err != nil {
		return nil, err
	}

	start := req.GetCursor()
	end := start + int64(req.GetLimit())

	messages, err := RedisDB.RedisGetMessagesByRoomID(ctx, roomID, start, end, req.GetReverse())
	if err != nil {
		return nil, err
	}

	respMessages := make([]*rpc.Message, 0)
	var counter int32 = 0
	var next int64 = 0
	hasMore := false
	for _, msg := range messages {
		// There are more message beyond the required limit
		if counter+1 > req.GetLimit() {
			hasMore = true
			next = end
			break
		}

		tempMessage := &rpc.Message{
			Chat:     req.GetChat(),
			Text:     msg.Message,
			Sender:   msg.Sender,
			SendTime: msg.Timestamp,
		}
		respMessages = append(respMessages, tempMessage)
		counter += 1
	}

	resp := rpc.NewPullResponse()
	resp.Messages = respMessages
	resp.Code, resp.Msg = 0, "success"
	resp.HasMore, resp.NextCursor = &hasMore, &next
	return resp, nil
}

// Retrieve roomID (string) of chat message
// a1:a2 equivalent to a2:a1 - perform checking
func getRoomID(chat string) (string, error) {
	lowercase := strings.ToLower(chat)
	// split a1:a2 into [a1, a2]
	senders := strings.Split(lowercase, ":")
	// Enforce one to one chat
	if len(senders) != 2 {
		err := fmt.Errorf("Invalid chat ID format: %s, Correct format: sender1:sender2", chat)
		return "", err
	}

	sender1, sender2 := senders[0], senders[1]
	var roomID string
	// Compare the senders alphabetically, ascending order
	if comp := strings.Compare(sender1, sender2); comp == 1 {
		roomID = fmt.Sprintf("%s:%s", sender2, sender1)
	} else {
		roomID = fmt.Sprintf("%s:%s", sender1, sender2)
	}

	return roomID, nil
}

func valSendReq(req *rpc.SendRequest) error {
	senders := strings.Split(req.Message.Chat, ":")
	if len(senders) != 2 {
		err := fmt.Errorf("Invalid chat ID format: %s, Correct format: sender1:sender2", req.Message.GetChat())
		return err
	}
	sender1, sender2 := senders[0], senders[1]

	// The message does not include the sender :o
	if req.Message.GetSender() != sender1 && req.Message.GetSender() != sender2 {
		err := fmt.Errorf("Sender %s not in the chat room", req.Message.GetSender())
		return err
	}

	return nil
}
