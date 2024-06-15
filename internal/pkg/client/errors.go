package client

type ClientError int

const (
	CLOSED_CLIENT      ClientError = 0
	CLIENT_BUFFER_FULL ClientError = 1
)

func (ClientError) Error() string {
	switch ClientError(CLOSED_CLIENT) {
	case CLOSED_CLIENT:
		return "Client is closed"
	case CLIENT_BUFFER_FULL:
		return "Client's buffer is full"
	default:
		return "Unknown client error"

	}

}
