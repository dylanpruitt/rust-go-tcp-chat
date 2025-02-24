
# Simple TCP Chat Server/Client
I've gone through the official tutorials for both Rust and Go, but wanted to work on a project to have hands-on experience with both. I thought making a simple chat app with both languages would be a great way to compare the two.

This repo builds off of two similar existing repos:  
Go TCP Server/Client (https://github.com/jeroendk/go-tcp-udp)  
Rust TCP Server/Client (https://github.com/EleftheriaBatsou/chat-app-client-server-rust)

### Overview
**Rust Server**  
The server creates a TCP listener for clients to connect to, and spawns a separate thread to read user input. Input read from this thread is broadcasted as a message to all clients, unless the user inputs `:quit`, which shuts down the server. Until the server shuts down, it infinitely loops to check for client connections, adding additional threads for each client connection to listen for client messages.

**Rust Client**  
Gets client username from stdin, and starts a nonblocking connection to the server. Spawns a separate thread to listen for messages sent from the server; the main thread reads user messages from stdin to send to the server until the user inputs `:quit`.  

**Go Server/Client**
The Go implementation is mostly the same, with two major exceptions: Go's `TCPListener.Accept()` function is blocking, so I had to make another separate goroutine to write messages to clients, and I use mutexes to manage shared state in both the server and the client.

### Run Instructions

To run the Rust client/server, open two terminals: one in `rust-client/` and one in `rust-server/`. Start the client and server by running `cargo run` in each terminal.  
To run the Go client/server, open two terminals: one in `go-client/` and one in `go-server/`. Start the client and server by running `go run main.go` in each terminal.

### Commands

- `:help` (**client**): prints a help message showing all valid commands.
- `:user [NEW_USERNAME]` (**client**): sets client username to `[NEW_USERNAME]`.
- `:quit` (**client/server**): terminates client/server connection and exits the application.
