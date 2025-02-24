package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
    "strings"
    "sync"
    "time"
)

// Listener/shutdown are read/written to across multiple goroutines, so they need a mutex.
type Server struct {
    mu sync.Mutex
    l *net.TCPListener // tcp listener
    s  bool        // flag to shutdown server
}

func (u *Server) Listener() *net.TCPListener {
    u.mu.Lock()
    defer u.mu.Unlock()
    return u.l
}

func (u *Server) IsShutdown() bool {
    u.mu.Lock()
    defer u.mu.Unlock()
    return u.s
}

func (u *Server) Shutdown() {
    u.mu.Lock()
    u.s = true
    u.l.Close()
    u.mu.Unlock()
}

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

    // Server uses a channel to communicate messages between goroutines
    messages := make(chan string)
    server := Server{l:listener, s: false}

    // Spawns a goroutine responsible for writing messages to all clients when they are received.
    go func() {
        for {
            if server.IsShutdown() {
                return
            }
                
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

    // Spawns a goroutine to read user input from stdin.
    go func() {
        reader := bufio.NewReader(os.Stdin)

        for {
            message, err := reader.ReadString('\n')
            if err != nil {
                fmt.Println("Error reading input:", err)
            }
            
            // If user inputs :quit, quit writing to the server and shutdown.
            if strings.TrimSpace(message) == ":quit" {
                server.Shutdown()
                return
            }

            messages <- message
        }
    }()
    
    // Loops to check for new connections, spawning a goroutine to handle each client connection for each one that joins.
	for {
		// Accept new connections
		conn, err := server.Listener().Accept()
        clients = append(clients, conn)
		if err != nil {
			fmt.Println(err)
            return
		}

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
            fmt.Println(fmt.Sprintf("closing connection with %s (%s)", username, conn.RemoteAddr().String()))
            messages <- fmt.Sprintf("%s disconnected.\n", username)
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
