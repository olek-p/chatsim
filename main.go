package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// grab the number of users to simulate
	var userCount int
	flag.IntVar(&userCount, "users", 4, "Number of concurrent chat users (minimum of 2)")
	flag.Parse()
	if userCount < 2 {
		fmt.Printf("invalid value \"%d\" for flag -users\n", userCount)
		flag.Usage()
		os.Exit(2)
	}

	fmt.Printf("Starting a chat sim with %d users\n", userCount)

	// create the right number of users
	users := []*User{}
	for i := 1; i <= userCount; i++ {
		newUser := User{
			ID:           fmt.Sprintf("user_%d", i),
			UserChannel:  make(chan *User),
			EventChannel: make(chan ChatEvent),
			OtherUsers:   make(map[string]*User),
			UserChats:    make(map[string]*ChatRoom),
		}

		// initialize the new user (set up channels etc.)
		go newUser.Start()

		// notify the new user of all users so far
		// + all users so far of the new user
		for _, user := range users {
			newUser.UserChannel <- user
			user.UserChannel <- &newUser
		}
		users = append(users, &newUser)
	}

	// allow the users to randomly use chat forever
	for {
	}
}
