package main

import (
	"encoding/json"
	"log"
)

type Message struct {
	Author *Client `json:"author"`
	Action string  `json:"action"`
	Body   string  `json:"body"`
	// Target Room ID
	Target string `json:"target"`
}

const (
	/*
			TextMessage
		{
			Author: 0x0031f (pointer to client),
			Action: "TextMessage",
			Body: "Hello, World!",
			Target: "room-1-id"
		}
	*/
	TextMessageAction = "TextMessage"

	/*
			JoinRoomMessage
		{
			Author: 0x0031f (pointer to client),
			Action: "JoinRoom",
			Body: "room-name-1",
			Target: nil
		}
	*/
	JoinRoomAction = "JoinRoom"

	/*
			LeaveRoomMessage
		{
			Author: 0x0031f (pointer to client),
			Action: "LeaveRoom",
			Body: "room-name-1",
			Target: nil
		}
	*/
	LeaveRoomAction = "LeaveRoom"

	/*
			RoomJoinedAction
		{
			Author: 0x0031f (pointer to client),
			Action: "JoinedRoom",
			Body: "room-name-1",
			Target: "room-id-1"
		}
	*/
	RoomJoinedAction = "JoinedRoom"
)

func (m *Message) encode() []byte {
	jsonMessage, err := json.Marshal(*m)
	if err != nil {
		log.Fatal("Failed to marshall json message")
	}
	return jsonMessage
}

func decodeMessage(jsonMessage []byte) (message Message) {
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Fatal(err)
		return
	}

	return message
}
