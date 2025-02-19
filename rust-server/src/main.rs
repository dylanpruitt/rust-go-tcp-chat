use std::io::{self, ErrorKind, Read, Write};
use std::net::TcpListener;
use std::sync::mpsc;
use std::thread;

const LOCAL: &str = "127.0.0.1:6000";
const MSG_SIZE: usize = 32;

fn sleep() {
	// Used to break up thread/server loops so they aren't constantly running.
    thread::sleep(::std::time::Duration::from_millis(100));
}

fn main() {
    let server = TcpListener::bind(LOCAL).expect("Listener failed to bind");
	// Allows the server to continually check for client messages.
    server.set_nonblocking(true).expect("failed to initialize non-blocking");
	// Vector of TcpStream connections to the server.
    let mut clients = vec![];

    let (tx, rx) = mpsc::channel::<String>();

	// The server spawns a separate thread to handle IO, allowing it to read user messages.
	// Unless the user specifies the server should quit (:quit), it sends the message out to all clients.
	let io_tx = tx.clone();
	thread::spawn(move || loop {
        let mut buff = String::new();
        io::stdin().read_line(&mut buff).expect("reading from stdin failed");
        let msg = buff.trim().to_string();
		// io_tx.send takes ownership of msg, so "if msg == :quit" is invalid after it.
		// I declare quitting to track whether the user sent ":quit" before msg is moved.
		let quitting:bool = if msg == ":quit" { true } else { false };
        if io_tx.send(msg).is_err() || quitting {
			break
		}
	});

    loop {
        if let Ok((mut socket, addr)) = server.accept() {
            println!("Client {} connected", addr);

            let tx = tx.clone();
            clients.push(socket.try_clone().expect("failed to clone client"));

			// Tracks the client's username.
			let mut username = String::new();

			// Spawns a separate thread for each client to listen for messages.
            thread::spawn(move || loop {
                let mut buff = vec![0; MSG_SIZE];

				// Reads messages from client if one was sent. Ends thread loop if unable to reach client.
                match socket.read_exact(&mut buff) {
                    Ok(_) => {
                        let msg = buff.into_iter().take_while(|&x| x != 0).collect::<Vec<_>>();
                        let msg = String::from_utf8(msg).expect("Invalid utf8 message");

						if msg.contains("user:") {
							// user:USERNAME messages tell the server to store the client's username.
							username = msg.strip_prefix("user:").unwrap().trim().to_string();
							println!("{} is user {:?}", addr, username);
						} else {
							// Print client message and who sent it.
							let message_with_sender: String = format!("{username}: {msg}");
							println!("{}", message_with_sender);
							tx.send(message_with_sender).expect("failed to send msg to rx");
						}
                    }, 
                    Err(ref err) if err.kind() == ErrorKind::WouldBlock => (),
                    Err(_) => {
                        println!("closing connection with {} ({})", username, addr);
                        break;
                    }
                }

                sleep();
            });
        }
		
		let mut shutdown_server:bool = false;
		// If the user inputs ':quit', exits the server loop. Otherwise, sends msg to all clients.
        if let Ok(msg) = rx.try_recv() {
			if msg == ":quit" {
				shutdown_server = true;
			} else {
				clients = clients.into_iter().filter_map(|mut client| {
					let mut buff = msg.clone().into_bytes();
					buff.resize(MSG_SIZE, 0);

					client.write_all(&buff).map(|_| client).ok()
				}).collect::<Vec<_>>();
			}
        }
		if shutdown_server { break }
        sleep();
    }

	println!("Shutdown server.");
}
