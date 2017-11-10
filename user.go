package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// User is our simulated chat user
type User struct {
	ID           string
	UserChannel  chan *User           // used to send presence notifications
	EventChannel chan ChatEvent       // used to send chat events
	OtherUsers   map[string]*User     // holds all other online users
	UserChats    map[string]*ChatRoom // holds all chat rooms this user is in
}

// Start initializes the chat user
func (u *User) Start() {
	for {
		select {
		case user := <-u.UserChannel:
			// only new online users pass through here
			// the result is to output the list of users currently "seen" by this user
			ids := u.handleNewUser(user)
			u.logMsg(fmt.Sprintf("I see %d users: %s", len(ids), strings.Join(ids, ", ")))
		case event := <-u.EventChannel:
			// this channel receives 3 types of events:
			switch event.Type {
			case ChatEventNew: // new chats are joined or acknowledged by the recipient user
				if isInChat := u.handleNewChat(event.ChatRoom); isInChat {
					u.logMsg(fmt.Sprintf("I joined chat %s", event.ChatRoom.ID))
				} else {
					u.logMsg(fmt.Sprintf("I see a new chat %s", event.ChatRoom.ID))
				}
			case ChatEventMessage: // new messages are only sent to users in chat rooms
				u.logMsg(fmt.Sprintf("I received message \"%s\" from chat %s", event.Message, event.ChatRoom.ID))
			case ChatEventClose: // closing a chat room is acknowledged by users in it and not in it alike
				u.handleChatClose(event.ChatRoom.ID)
				u.logMsg(fmt.Sprintf("I acknowledged closing of chat %s", event.ChatRoom.ID))
			}
		case <-time.After(time.Second * 1):
			// every second, one of 3 actions may be initiated, depending on a pseudo-random number generator
			switch ChatEventType(rand.Intn(10)) {
			case ChatEventNew: // creating a new chat room with randomly selected users - only if one with the same users doesn't exist
				if event, err := u.NewChat(); err != nil {
					u.logMsg(fmt.Sprintf("I wanted to find some single-serving friends, but an error occurred: %s", err.Error()))
				} else {
					u.logMsg(fmt.Sprintf("I created a new chatroom: %s", event.ChatRoom.ID))
				}
			case ChatEventMessage: // sending a message to a chat room this user is in (if they are in any)
				if event, err := u.NewMessage(); err != nil {
					u.logMsg(fmt.Sprintf("I wanted to talk about fight club, but an error occurred: %s", err.Error()))
				} else {
					u.logMsg(fmt.Sprintf("I sent a message \"%s\" to chat %s", event.Message, event.ChatRoom.ID))
				}
			case ChatEventClose: // closing a randomly selected chat room this user is in (if they are in any)
				if event, err := u.CloseChat(); err != nil {
					u.logMsg(fmt.Sprintf("I wanted to destroy something beautiful, but an error occurred: %s", err.Error()))
				} else {
					u.logMsg(fmt.Sprintf("I closed chat %s", event.ChatRoom.ID))
				}
				// default: no action
			}
		}
	}
}

// logMsg is a helper method, formatting and logging the passed msg to the standard output
func (u *User) logMsg(msg string) {
	fmt.Printf("[%-15s %7s] %s\n", time.Now().Format("15:04:05.999999"), u.ID, msg)
}

// otherUserIDs is a boilerplate function to get map's keys as a slice
func (u *User) otherUserIDs() []string {
	ids := make([]string, 0, len(u.OtherUsers))
	for id := range u.OtherUsers {
		ids = append(ids, id)
	}

	return ids
}

// getRandomChat selects a random chat from this user's chat rooms, or error if none exist
func (u *User) getRandomChat() (*ChatRoom, error) {
	chatCount := len(u.UserChats)
	if chatCount == 0 {
		return &ChatRoom{}, errors.New("No chats available")
	}

	// pick a random chat room this user is in
	chatIDs := make([]string, 0, len(u.UserChats))
	for chatID := range u.UserChats {
		chatIDs = append(chatIDs, chatID)
	}
	randomIndex := rand.Intn(chatCount)

	return u.UserChats[chatIDs[randomIndex]], nil
}

// handleNewUser adds the newUser to the list of this user's OtherUsers
// this function is idempotent
func (u *User) handleNewUser(newUser *User) []string {
	u.OtherUsers[newUser.ID] = newUser

	return u.otherUserIDs()
}

// handleNewChat adds the newChat to the list of this user's UserChats,
// provided they are in it
func (u *User) handleNewChat(newChat *ChatRoom) bool {
	isInChat := false
	for _, userID := range newChat.UserIDs {
		if u.ID == userID {
			isInChat = true
		}
	}
	if isInChat {
		u.UserChats[newChat.ID] = newChat
	}

	return isInChat
}

// handleChatClose deletes the chat identified by chatID from the list
// of this user's UserChats. This function is idempotent
func (u *User) handleChatClose(chatID string) {
	delete(u.UserChats, chatID)
}

// NewChat creates a new chat room and notifies all other users about it
func (u *User) NewChat() (ChatEvent, error) {
	// total size of new chat, including creator
	size := rand.Intn(len(u.OtherUsers)) + 2
	// select size-1 other users randomly
	perm := rand.Perm(size - 1)
	otherUserIDs := u.otherUserIDs()

	// translate random numbers to user ID's
	ids := make([]string, 0, size)
	for _, v := range perm {
		ids = append(ids, otherUserIDs[v])
	}
	// include creator ID
	ids = append(ids, u.ID)
	// make chatroom ID's deterministic
	sort.Strings(ids)
	chatroomID := strings.Join(ids, "_")

	// only start a new chat if it's unique
	if _, found := u.UserChats[chatroomID]; found {
		return ChatEvent{}, fmt.Errorf("Duplicated chat ID: %s", chatroomID)
	}

	chatroom := ChatRoom{
		ID:      strings.Join(ids, "_"),
		UserIDs: ids,
	}

	// notify ALL other users of the new chat room
	event := ChatEvent{u, &chatroom, "New chat created", ChatEventNew, time.Now()}
	for _, user := range u.OtherUsers {
		go func(user *User) {
			user.EventChannel <- event
		}(user)
	}
	// save the new chat in creator's cache
	u.UserChats[chatroomID] = &chatroom

	return event, nil
}

// NewMessage sends a message from this user to one if their chat rooms
func (u *User) NewMessage() (ChatEvent, error) {
	selectedChat, err := u.getRandomChat()
	if err != nil {
		return ChatEvent{}, err
	}

	// prepare the message and send it to other users in the same chat room
	ts := time.Now()
	event := ChatEvent{
		User:     u,
		ChatRoom: selectedChat,
		Message:  fmt.Sprintf("Hello from %s at %s", u.ID, ts.Format("15:04:05")),
		Type:     ChatEventMessage,
		Ts:       ts,
	}
	for _, userID := range selectedChat.UserIDs {
		if userID != u.ID { // omit the message author
			go func(userID string) {
				u.OtherUsers[userID].EventChannel <- event
			}(userID)
		}
	}

	return event, nil
}

// CloseChat closes one of this user's chat rooms
func (u *User) CloseChat() (ChatEvent, error) {
	selectedChat, err := u.getRandomChat()
	if err != nil {
		return ChatEvent{}, err
	}

	// generate the close event and send it to ALL users
	event := ChatEvent{
		User:     u,
		ChatRoom: selectedChat,
		Message:  "",
		Type:     ChatEventClose,
		Ts:       time.Now(),
	}
	for userID := range u.OtherUsers {
		go func(userID string) {
			u.OtherUsers[userID].EventChannel <- event
		}(userID)
	}

	delete(u.UserChats, selectedChat.ID)

	return event, nil
}
