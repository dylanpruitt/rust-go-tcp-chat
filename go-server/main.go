package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
    "strings"
    "time"
)

func main() {

	const TcpAddr = "127.0.0.1:6000"
    clients := make([]net.Conn, 0)

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

    messages := make(chan string)

    go func() {
        for {
            select {
            case message := <-messages:
                // Print the data read from the connection to the terminal
                fmt.Print("> ", message)
                for _, client := range clients {
                    client.Write([]byte(message))
                }
            case <-time.After(100 * time.Millisecond):
                // Continue to wait for messages
                continue
            }
        }
    }()
    
	for {
		// Accept new connections
        fmt.Println("block")
		conn, err := listener.Accept()
        fmt.Println("block2")
        clients = append(clients, conn)
		if err != nil {
			fmt.Println(err)
            continue
		}
        
        fmt.Println("?")
		// Handle new connections in a Goroutine for concurrency
		go handleConnection(conn, messages)
        
	}
}

func handleConnection(conn net.Conn, messages chan<- string) {
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
		

        if strings.Contains(message, ":user ") {
            // user:USERNAME messages tell the server to store the client's username.
            clientIP := conn.RemoteAddr().String()
            oldUsername := username
            username = strings.TrimSpace(strings.TrimPrefix(message, ":user "))
            fmt.Println(clientIP, "is user", username)
            if oldUsername != "" {
                // If not the client initially setting their username, send a message with updated username.
                messages <- fmt.Sprintf("%s changed username to %s\n", oldUsername, username)
            } else {
                // Sends a message for client initially connecting.
                messages <- fmt.Sprintf("%s joined the server\n", username)
            }
        } else {
            // Print client message and who sent it.
            messageWithSender := fmt.Sprintf("%s: %s", username, message)
            fmt.Println(messageWithSender)
            messages <- messageWithSender
        }
	}
}
