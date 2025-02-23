package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
    "strings"
)

func main() {

	const TcpAddr = "127.0.0.1:6000"

	// Resolve the string address to a TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp4", TcpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start listening for TCP connections on the given address
	listener, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		// Accept new connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		// Handle new connections in a Goroutine for concurrency
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
    username := ""

	for {
		// Read from the connection untill a new line is send
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

        message := string(data)
		// Print the data read from the connection to the terminal
		fmt.Print("> ", message)

        if strings.Contains(message, ":user ") {
            // user:USERNAME messages tell the server to store the client's username.
            // TODO check for oldUsername --> oldUsername := username
            // TODO send welcome messsage or user change message
            clientIP := conn.RemoteAddr().String()
            username = strings.TrimSpace(strings.TrimPrefix(message, ":user "))
            fmt.Println(clientIP, "is user", username)
            conn.Write([]byte("Hello TCP Client\n"))
        } else {
            // Write back the same message to the client
            fmt.Print(username)
            conn.Write([]byte("Hello TCP Client\n"))
        }
	}
}
