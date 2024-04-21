# TCPSession

## Description

Go websocket server for sending messages between clients in a session. Fork of [WsSession](https://github.com/Aaron-json/WsSession). If you need this functionality in a web browser consider using [WsSession](https://github.com/Aaron-json/WsSession), since Websockets have much better support than raw TCP connections on the browser.

## Table of Contents

- [Usage](#usage)
- [Codes](#codes)
- [Contributing](#contributing)

## Usage

1. Fork the source code to get your own copy and compile it.

2. Start the server by running the executable.

3. Create a TCP connection to the server. The server will listen on the connection for the first message. The first message from the server be MUST one of two types, joining a session or creating a new session. To create a new session send a single byte with the value 0. To join a session send a single byte with the value 1 followed by the code of the session you want to join.

    *Note: The server only accepts message of the right length, messages of unexpected length will NOT be accepted.*


4. If creating a new session is successful, the server will respond with a message containing status 0 (the first byte) followed by the session code. If joining an existing session is successful, the session will respond with byte 0 only. For all response codes, refer to [Response](#response).

5. Use this code to connect other clients to the same session and share messages.

## Codes

### Response

	SUCCESS            = 0
	ERROR              = 1
	SESSION_NOT_FOUND  = 2
	SESSION_FULL       = 3
	SERVER_FULL        = 4
	INVALID_ACTION     = 5

### Request
	CREATE_SESSION     = 0
	JOIN_SESSION       = 1

## Contributing

If you would like to contribute or outline issues and potential improvements, feel free to raise an issue or create a pull request.