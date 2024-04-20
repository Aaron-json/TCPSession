package controllers

const (
	MEMBER_JOINED = 0
	MEMBER_LEFT   = 1
)

// Sends a message to all other clients in the session. Handles synchronization
// using a read lock on the session.
func Broadcast(Sender Member, msg []byte) {
	Sender.Session.mu.RLock()
	defer Sender.Session.mu.RUnlock()
	BroadcastUnsync(Sender, msg)
}

// Broadcasts message to members without any synchronization. The caller is responsible
// for handling synchronization. Most callers should consider using Broadcast() unless
// they need to broadcast within an existing lock.
func BroadcastUnsync(Sender Member, data []byte) {
	for _, member := range Sender.Session.Members {
		if member.Id != Sender.Id {
			member.Client.Send(data)
		}
	}
}
