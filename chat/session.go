package chat

import "net"

// Session contains info about a connection between client and server.
type Session struct {
	conn *net.Conn //self
}


func (s *Session) handleSession() {
	
}
