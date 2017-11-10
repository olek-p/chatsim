package main

import "time"

type (
	// ChatRoom represents a simulated chat room
	ChatRoom struct {
		ID      string   // combination of ID's of users in it
		UserIDs []string // kind of redundant, but left for clarity
	}
	// ChatEvent carries chat events between users
	ChatEvent struct {
		User     *User         // user triggering the event
		ChatRoom *ChatRoom     // chat room the event affects
		Message  string        // optional message
		Type     ChatEventType // type of event
		Ts       time.Time     // event timestamp
	}
	// ChatEventType is just a wrapper for int
	ChatEventType int
)

const (
	// ChatEventNew is an event sent on chat room creation
	ChatEventNew ChatEventType = iota
	// ChatEventMessage is sent when a chat message is sent to a chat room
	ChatEventMessage
	// ChatEventClose marks chat room closing by one of the users in it
	ChatEventClose
)
