use std::io::{self, ErrorKind, Read, Write};
use std::net::TcpStream;
use std::sync::mpsc::{self, TryRecvError};
use std::thread;
use std::time::Duration;

const LOCAL: &str = "127.0.0.1:6000";
const MSG_SIZE: usize = 32;

fn main() {
    let mut username = String::new();
    print!("Enter a username:");
    // stdout is usually line-buffered, makes sure the print macro above shows BEFORE inputting username.
    let _ = io::stdout().flush();
    io::stdin().read_line(&mut username).expect("reading from stdin failed");

    let mut client = TcpStream::connect(LOCAL).expect("Stream failed to connect");
    client.set_nonblocking(true).expect("failed to initiate non-blocking");

    let (tx, rx) = mpsc::channel::<String>();

    // TODO: revisit unwrap, not sure if there's a better thing to do for error handling here.
    tx.send(format!("user:{username}")).unwrap();

    thread::spawn(move || loop {
        let mut buff = vec![0; MSG_SIZE];
        match client.read_exact(&mut buff) {
            Ok(_) => {
                let msg_bytes: Vec<u8> = buff.into_iter().take_while(|&x| x != 0).collect();
                let msg_str: String = String::from_utf8(msg_bytes).expect("message should contain valid utf8 bytes");
                println!("message recv {:?}", msg_str);
            },
            Err(ref err) if err.kind() == ErrorKind::WouldBlock => (),
            Err(_) => {
                println!("connection with server was severed");
                break;
            }
        }

        match rx.try_recv() {
            Ok(msg) => {
                let mut buff = msg.clone().into_bytes();
                buff.resize(MSG_SIZE, 0);
                client.write_all(&buff).expect("writing to socket failed");
                if !msg.contains("user:") {
                    println!("message sent {:?}", msg);
                }
            }, 
            Err(TryRecvError::Empty) => (),
            Err(TryRecvError::Disconnected) => break
        }

        thread::sleep(Duration::from_millis(100));
    });

    println!("Write a Message:");
    loop {
        let mut buff = String::new();
        io::stdin().read_line(&mut buff).expect("reading from stdin failed");
        let msg = buff.trim().to_string();
        if msg == ":quit" || tx.send(msg).is_err() {break}
    }
    println!("Disconnected from server.");
}

// To run this program you need to open 2 terminals. One for the client and one for the server. 
// In the server run `cargo run`. 
// Then do the same in your client. And this time you should see a message, `write a message`. 
// Type something and then you should see that in the server. 
// If you type ':quit' then the program will quit...
