# One-To-One Chat

The One-To-One Chat project implements 1:1 message to any active username via websocket simply.

- Features
  - 1:1 message
  - chatroom broadcast
  - username unique validation

## Demo

<img src="./asset/demo.gif" width="80%"/>

### Execute Project

````bash
git clone https://github.com/MikeHsu0618/go-websocket-demo.git
cd otochat
docker compose up
````

### Available Services

* starts a websocket server on localhost port 8080 (by default).
* ws://localhost:8080/ws?username={username}

| Method | Path                    | Usage                                  |
|--------|-------------------------|----------------------------------------|
| GET    | /ws?username={username} | join the chatroom with unique username |


### TODO

- [ ] workflow chart
- [ ] consistent storage
- [ ] multiple room management
- [ ] testing
- [ ] implement front-end 
