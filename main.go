package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// BroadcastMessage broadcast received message to all clients currently connected.
func BroadcastMessage(conn net.Conn, clients []net.Conn) {
	for {
		// will listen for message to process ending in newline (\n)
		message, _ := bufio.NewReader(conn).ReadString('\n')
		// send new string back to client
		for _, cli := range clients {
			cli.Write([]byte(message))
		}
	}
}

// ShowMenu show menu when a client logon.
/* func ShowMenu(conn net.Conn) int {

	var data []string
	data = data.append(data, "1:注册\n")
	data = data.append("2:登陆\n")
	//data=data.append("")
	conn.Write([]byte(data))
} */

func main() {
	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Printf("listen failed: %v", err)
		os.Exit(1)
	}

	var clients []net.Conn
	//num := 0

	// accept connection on port
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("accept failed: %v", err)
			continue
		}
		//fmt.Println(conn.LocalAddr())
		//fmt.Println(conn.RemoteAddr())
		clients = append(clients, conn)

		go func() {
			//ShowMenu(conn)
			BroadcastMessage(conn, clients)
		}()
	}
}
