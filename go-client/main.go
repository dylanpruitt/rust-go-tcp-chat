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

// Username is read/written to across multiple goroutines, so it needs a mutex.
type Username struct {
    mu sync.Mutex
    v  string
}

func (u *Username) Get() string {
    u.mu.Lock()
    defer u.mu.Unlock()
    return u.v
}

func (u *Username) Set(newUsername string) {
    u.mu.Lock()
    u.v = newUsername
    u.mu.Unlock()
}

func main() {

    const TcpAddr = "127.0.0.1:6000"

	// Resolve the string address to a TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp4", TcpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

    fmt.Print("Enter a username:");
    reader := bufio.NewReader(os.Stdin)
	un, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
    username := Username{v: strings.TrimSpace(un)}

	// Connect to the address with tcp
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

    fmt.Println("Write a Message:")
    messages := make(chan string)

    go readFromServer(conn, &username)
    go writeUserMessageTo(conn, messages)
    
    // Send client username to the server
    fmt.Println("> setting username as", username.Get())
    messages <- fmt.Sprintf(":user %s\n", username.Get())

    for {
        message, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("Error reading input:", err)
        }
        
        // If client inputs :help, quit writing to the server.
        if strings.TrimSpace(message) == ":quit" {
            close(messages)
            return
        }
        // If client inputs :help, print help message and do not send to the server.
        if strings.TrimSpace(message) == ":help" {
            fmt.Println("Chat App Commands:")
            fmt.Println(":help - display this message")
            fmt.Println(":user [USER] - change username")
            fmt.Println(":quit - disconnect from server")
            continue;
        }

        if strings.Contains(message, ":user") {
            // Client will not send whitespace/empty usernames to the server.
            messageUsername := strings.TrimSpace(strings.TrimPrefix(message, ":user"))
            // Displays help message if :user command is used incorrectly.
            if messageUsername == "" {
                fmt.Println("INVALID USE OF :user COMMAND");
                fmt.Println("Type ':user [USERNAME]' to set your username (ex. ':user Ringo')");
                continue
            }
        }
        
        // Send a message to the server
        if strings.Contains(message, ":user ") {
            newUsername := strings.TrimSpace(strings.TrimPrefix(message, ":user "))
            username.Set(newUsername)
            fmt.Println("> setting username as", newUsername)
        } else {
            fmt.Println(fmt.Sprintf("> sent message \"%s\"", strings.TrimSpace(message)))
        }
        messages <- message
    }
    
    fmt.Println("Disconnected from server.")
}

func readFromServer(conn net.Conn, username *Username) {
    defer conn.Close()

    for {
        // Read from the connection until a new line is send
        data, err := bufio.NewReader(conn).ReadString('\n')
        if err != nil {
            fmt.Println(err)
            return
        }

        // To avoid printing messages again after sending, only displays messages if they aren't from this client.
        message := string(data)
        usernamePrefix := fmt.Sprintf("%s:", username.Get())
        userJoinedMessage := fmt.Sprintf("%s joined the server.", username.Get())
        if !strings.HasPrefix(message, usernamePrefix) && message != userJoinedMessage {
            fmt.Print(message)
        } 
    }
}

func writeUserMessageTo(conn net.Conn, messages <-chan string) {
    for {
        select {
        case message := <-messages:
            // Send a message to the server
            _, err := conn.Write([]byte(message))
            if err != nil {
                fmt.Println(err)
                return
            }
        case <-time.After(100 * time.Millisecond):
            // Continue to wait for messages
            continue
        }
    }
}