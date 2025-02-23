package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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

	// Send a message to the server
	_, err = conn.Write([]byte("Hello TCP Server\n"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
    
    reader.ReadString('\n')
}

func readFromServer(conn net.Conn) {
    defer conn.Close()

    for {
        // Read from the connection untill a new line is send
        data, err := bufio.NewReader(conn).ReadString('\n')
        if err != nil {
            fmt.Println(err)
            return
        }

        // Print the data read from the connection to the terminal
        fmt.Print("> ", string(data))
    }
}