use aether_log_client::AetherClient;
use std::{env, error::Error, io};

fn main() -> Result<(), Box<dyn Error>> {
    let mut args = env::args().skip(1);
    let message = args.next().unwrap_or_else(|| "hello from rust".into());
    let address = args.next().unwrap_or_else(|| "127.0.0.1:8080".into());
    if args.next().is_some() {
        return Err(io::Error::new(
            io::ErrorKind::InvalidInput,
            "usage: cargo run --example send -- [message] [hub-address]",
        )
        .into());
    }
    let mut client = AetherClient::connect(address)?;
    client.send(message.as_bytes())?;
    client.close()?;
    println!("TCP write completed; Hub persistence is not acknowledged.");
    Ok(())
}
