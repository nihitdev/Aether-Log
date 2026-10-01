//! Synchronous Aether protocol v1 emitter with no external dependencies.
//!
//! Successful sends mean local TCP writes, not durable Hub acceptance.
//! There are no acknowledgements or automatic retries.

use std::io::{self, Write};
use std::net::{Shutdown, TcpStream, ToSocketAddrs};
use std::time::Duration;

const MAGIC: u32 = 0xAE74_4552;
const VERSION: u16 = 1;
const TYPE_DATA: u8 = 1;
const HEADER_SIZE: usize = 11;

/// Connection and send limits. Defaults match the Hub's 1 MiB payload limit.
#[derive(Debug, Clone)]
pub struct ClientOptions {
    /// Timeout for each resolved address attempted; DNS resolution is not bounded.
    pub connect_timeout: Duration,
    /// Socket write timeout, per blocking write rather than per complete frame.
    pub write_timeout: Duration,
    /// Maximum record bytes, independently limited by the v1 u32 wire length.
    /// Zero permits only empty records.
    pub max_payload: usize,
}

impl Default for ClientOptions {
    fn default() -> Self {
        Self {
            connect_timeout: Duration::from_secs(5),
            write_timeout: Duration::from_secs(5),
            max_payload: 1024 * 1024,
        }
    }
}

/// One reusable TCP connection. Sending requires exclusive mutable access.
/// Dropping the client releases its socket without requiring an explicit close.
#[derive(Debug)]
pub struct AetherClient {
    stream: Option<TcpStream>,
    max_payload: usize,
}

impl AetherClient {
    /// Connect using five-second connection/write timeouts and a 1 MiB limit.
    pub fn connect(address: impl ToSocketAddrs) -> io::Result<Self> {
        Self::connect_with_options(address, ClientOptions::default())
    }

    /// Try resolved addresses in order. All timeouts must be nonzero.
    pub fn connect_with_options(
        address: impl ToSocketAddrs,
        options: ClientOptions,
    ) -> io::Result<Self> {
        if options.connect_timeout.is_zero() || options.write_timeout.is_zero() {
            return Err(io::Error::new(
                io::ErrorKind::InvalidInput,
                "connection and write timeouts must be nonzero",
            ));
        }
        let mut last_error = io::Error::new(
            io::ErrorKind::InvalidInput,
            "address resolved to no socket addresses",
        );
        for address in address.to_socket_addrs()? {
            match TcpStream::connect_timeout(&address, options.connect_timeout) {
                Ok(stream) => {
                    stream.set_write_timeout(Some(options.write_timeout))?;
                    return Ok(Self {
                        stream: Some(stream),
                        max_payload: options.max_payload,
                    });
                }
                Err(error) => last_error = error,
            }
        }
        Err(last_error)
    }

    /// Send one Data frame without appending a newline.
    ///
    /// Payload validation happens before any write. Validation errors preserve
    /// the connection. A write error may mean partial/complete transmission; it
    /// invalidates the connection to prevent sending after a truncated frame.
    /// Create a new client to reconnect; retrying may duplicate the record.
    pub fn send(&mut self, payload: &[u8]) -> io::Result<()> {
        let stream = self.stream.as_mut().ok_or_else(|| {
            io::Error::new(io::ErrorKind::NotConnected, "Aether client is closed")
        })?;
        let header = encode_header(payload.len(), self.max_payload)?;
        if let Err(error) = write_frame(stream, &header, payload) {
            if let Some(stream) = self.stream.take() {
                let _ = stream.shutdown(Shutdown::Both);
            }
            return Err(error);
        }
        Ok(())
    }

    /// Shut down both socket directions and release the connection.
    /// Repeated calls succeed. The client remains closed even if shutdown fails.
    /// This is not a persistence acknowledgement or a Hub drain operation.
    pub fn close(&mut self) -> io::Result<()> {
        match self.stream.take() {
            Some(stream) => stream.shutdown(Shutdown::Both),
            None => Ok(()),
        }
    }
}

fn encode_header(length: usize, limit: usize) -> io::Result<[u8; HEADER_SIZE]> {
    let wire_length = u32::try_from(length).map_err(|_| {
        io::Error::new(io::ErrorKind::InvalidInput, "payload exceeds v1 u32 length")
    })?;
    if length > limit {
        return Err(io::Error::new(
            io::ErrorKind::InvalidInput,
            format!("payload length {length} exceeds configured limit {limit}"),
        ));
    }
    let mut header = [0; HEADER_SIZE];
    header[..4].copy_from_slice(&MAGIC.to_be_bytes());
    header[4..6].copy_from_slice(&VERSION.to_be_bytes());
    header[6] = TYPE_DATA;
    header[7..].copy_from_slice(&wire_length.to_be_bytes());
    Ok(header)
}

fn write_frame(
    writer: &mut impl Write,
    header: &[u8; HEADER_SIZE],
    payload: &[u8],
) -> io::Result<()> {
    writer.write_all(header)?;
    writer.write_all(payload)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Read;
    use std::net::TcpListener;
    use std::thread;

    #[test]
    fn exact_header() {
        assert_eq!(
            encode_header(0x0102_0304, usize::MAX).unwrap(),
            [0xae, 0x74, 0x45, 0x52, 0, 1, 1, 1, 2, 3, 4]
        );
    }

    #[test]
    fn empty_and_normal_payloads() {
        for payload in [b"".as_slice(), b"hi", b"\0binary\xff\n"] {
            let mut wire = Vec::new();
            let header = encode_header(payload.len(), 100).unwrap();
            write_frame(&mut wire, &header, payload).unwrap();
            assert_eq!(&wire[..7], &[0xae, 0x74, 0x45, 0x52, 0, 1, 1]);
            assert_eq!(&wire[7..11], &(payload.len() as u32).to_be_bytes());
            assert_eq!(&wire[11..], payload);
        }
    }

    #[test]
    fn configured_payload_limit() {
        assert!(encode_header(0, 0).is_ok());
        assert!(encode_header(3, 3).is_ok());
        assert_eq!(
            encode_header(4, 3).unwrap_err().kind(),
            io::ErrorKind::InvalidInput
        );
    }

    #[test]
    #[cfg(target_pointer_width = "64")]
    fn oversized_wire_length_without_allocating() {
        assert!(encode_header(u32::MAX as usize, usize::MAX).is_ok());
        assert_eq!(
            encode_header(u32::MAX as usize + 1, usize::MAX)
                .unwrap_err()
                .kind(),
            io::ErrorKind::InvalidInput
        );
    }

    struct PartialWriter {
        bytes: Vec<u8>,
        interrupted: bool,
    }

    impl Write for PartialWriter {
        fn write(&mut self, bytes: &[u8]) -> io::Result<usize> {
            if !self.interrupted {
                self.interrupted = true;
                return Err(io::ErrorKind::Interrupted.into());
            }
            let length = bytes.len().min(2);
            self.bytes.extend_from_slice(&bytes[..length]);
            Ok(length)
        }
        fn flush(&mut self) -> io::Result<()> {
            Ok(())
        }
    }

    #[test]
    fn partial_and_interrupted_writes() {
        let mut writer = PartialWriter {
            bytes: Vec::new(),
            interrupted: false,
        };
        let header = encode_header(5, 5).unwrap();
        write_frame(&mut writer, &header, b"hello").unwrap();
        assert_eq!(writer.bytes, [header.as_slice(), b"hello"].concat());
    }

    #[test]
    fn zero_write_returns_error() {
        assert_eq!(
            write_frame(&mut &mut [][..], &encode_header(0, 0).unwrap(), b"")
                .unwrap_err()
                .kind(),
            io::ErrorKind::WriteZero
        );
    }

    #[test]
    fn zero_timeouts_rejected() {
        for options in [
            ClientOptions {
                connect_timeout: Duration::ZERO,
                ..ClientOptions::default()
            },
            ClientOptions {
                write_timeout: Duration::ZERO,
                ..ClientOptions::default()
            },
        ] {
            assert_eq!(
                AetherClient::connect_with_options("127.0.0.1:1", options)
                    .unwrap_err()
                    .kind(),
                io::ErrorKind::InvalidInput
            );
        }
    }

    #[test]
    fn write_failure_invalidates_connection() {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let mut client = AetherClient::connect(listener.local_addr().unwrap()).unwrap();
        let (_peer, _) = listener.accept().unwrap();
        client
            .stream
            .as_ref()
            .unwrap()
            .shutdown(Shutdown::Write)
            .unwrap();
        assert!(client.send(b"cannot write").is_err());
        assert!(client.stream.is_none());
        assert_eq!(
            client.send(b"again").unwrap_err().kind(),
            io::ErrorKind::NotConnected
        );
        client.close().unwrap();
    }

    #[test]
    fn multiple_sends_on_local_connection_and_close() {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let address = listener.local_addr().unwrap();
        let options = ClientOptions {
            max_payload: 5,
            ..ClientOptions::default()
        };
        let mut client = AetherClient::connect_with_options(address, options).unwrap();
        assert_eq!(
            client.stream.as_ref().unwrap().write_timeout().unwrap(),
            Some(Duration::from_secs(5))
        );
        let reader = thread::spawn(move || {
            let (mut stream, _) = listener.accept().unwrap();
            stream
                .set_read_timeout(Some(Duration::from_secs(5)))
                .unwrap();
            let mut wire = Vec::new();
            stream.read_to_end(&mut wire).unwrap();
            wire
        });
        assert_eq!(
            client.send(b"too long").unwrap_err().kind(),
            io::ErrorKind::InvalidInput
        );
        for payload in [b"first".as_slice(), b"", b"last"] {
            client.send(payload).unwrap();
        }
        client.close().unwrap();
        client.close().unwrap();
        assert_eq!(
            client.send(b"closed").unwrap_err().kind(),
            io::ErrorKind::NotConnected
        );
        let wire = reader.join().unwrap();
        let expected = [
            &[0xae, 0x74, 0x45, 0x52, 0, 1, 1, 0, 0, 0, 5][..],
            b"first",
            &[0xae, 0x74, 0x45, 0x52, 0, 1, 1, 0, 0, 0, 0],
            &[0xae, 0x74, 0x45, 0x52, 0, 1, 1, 0, 0, 0, 4],
            b"last",
        ]
        .concat();
        assert_eq!(wire, expected);
    }
}
