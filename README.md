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

3. Create a TCP connection to the server. When creating a session you only need to send a single byte that represents the operation as outlines in the [Response Codes](#response) section.

4. If creating a new session is successful, the server will respond with a 0 as the first byte followed by the session code. If joining an existing session is successful, the session will respond with a single byte indicating the status of the operation. For all response codes, refer to [Response Codes](#response).

5. After this handshake, the client can now send and receive data.

## Codes

### Request
	CREATE_SESSION     = 0
	JOIN_SESSION       = 1 (followed by the session code when joining a session)

### Response
	SUCCESS            = 0 (followed by the session code when creating a session)
	ERROR              = 1
	SESSION_NOT_FOUND  = 2
	SESSION_FULL       = 3
	SERVER_FULL        = 4
	INVALID_ACTION     = 5

## Contributing
If you would like to contribute ideas or outline issues and potential improvements, feel free to raise an issue or create a pull request.
