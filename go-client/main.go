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

    fmt.Print("Enter a username:");
    reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	// Connect to the address with tcp
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

    // Send client username to the server
	_, err = conn.Write([]byte(fmt.Sprintf(":user %s", username)))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
    }
    
    go readFromServer(conn)
    
    writeUserMessageTo(conn)
    fmt.Println("Disconnected from server.")
}

func readFromServer(conn net.Conn) {
    defer conn.Close()

    for {
        // Read from the connection until a new line is send
        data, err := bufio.NewReader(conn).ReadString('\n')
        if err != nil {
            fmt.Println(err)
            return
        }

        // Print the data read from the connection to the terminal
        fmt.Print(string(data))
    }
}

func writeUserMessageTo(conn net.Conn) {
    fmt.Println("Write a Message:")
    reader := bufio.NewReader(os.Stdin)
    
    for {
        message, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("Error reading input:", err)
        }
        
        // If client inputs :help, quit writing to the server.
        if strings.TrimSpace(message) == ":quit" { return }
        // If client inputs :help, print help message and do not send to the server.
        if strings.TrimSpace(message) == ":help" {
            fmt.Println("Chat App Commands:")
            fmt.Println(":help - display this message")
            fmt.Println(":user [USER] - change username")
            fmt.Println(":quit - disconnect from server")
            continue;
        }
        
        // Send a message to the server
        _, err = conn.Write([]byte(message))
        if err != nil {
            fmt.Println(err)
            return
        } else {
            if strings.Contains(message, ":user ") {
                newUsername := strings.TrimSpace(strings.TrimPrefix(message, ":user "))
                fmt.Println("> setting username as", newUsername)
            } else {
                fmt.Println(fmt.Sprintf("> sent message \"%s\"", strings.TrimSpace(message)))
            }
        }
    }
	
}