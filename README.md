# chatsim
chatsim is a simple multi-user chat simulation, written in Go. Its features include:
- users work asynchronously
- ability to specify the number of users
- each user has a unique uid
- each user knows of all other users
- each user knows of all active chat rooms
- there is not one central place holding the chat rooms; instead, the users control the rooms asynchronously
- each user randomly performs one of the following actions:
  - starts a chat with an arbitrary group of users
  - sends a message to a chat
  - ends a chat they are in

Some other choices I made:
- `User` and chat-related structs are exported for the sake of readability
- only standard library packages are used

## How it works
The application accepts a command-line flag `-users`, with value of type integer and a minimum of 2, specifying the number of simulated chat users. If not provided, **4 users are simulated**.
Upon start, the application creates the appropriate number of users, each in their own goroutine, each with two channels to communicate:
- `UserChannel` is used to pass information on new users appearing online
- `EventChannel` is used to pass information on chat room events, i.e. creation, messages, and deletion

Every second each user has a 30% chance of performing one of the actions:
- creating a chat room, randomly selecting users in it from the list of all online users, and then notifying everybody else of the new chat room. The chat room is not created if another already exists with the same group of users in it.
- sending a message to a room they are in (delivered to all the users in that room). This fails if the user is not in any chat room.
- closing a random chat room they are in (if there is one), and notifying all users of that

The output is logged to the console. It is pretty verbose and details each action from perspectives of each of the users involved.
The application goes on forever until Ctrl-C is pressed.

## Problems
There are some problems which should be addressed should this application be further developed:
- due to the decentralized nature of chat rooms storage, there is a potential race condition in the way chat rooms are created. It is possible that around the same point in time, two independent users create identical chat rooms, and start notifying users in them before themselves get notified of the other user's chat room. Similarly, one user may be sending a message to a chat room while another may be deleting it. This may lead to memory leaks.
- that empty `for` loop at the end of `main()` is ugly
- I haven't profiled the application e.g. to optimize the number of allocations or find any leaking goroutines
- there is no test coverage
