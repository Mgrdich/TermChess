//! Integration tests exercising the HTTP surface of the updater against a
//! minimal local test server. Ports the Go `httptest`-based tests.

use std::collections::HashMap;
use std::io::{Read, Write};
use std::net::TcpListener;
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::Duration;

use updater::{Client, Context, UpdaterError};

/// A request recorded by the test server.
#[derive(Default, Clone)]
struct RecordedRequest {
    path: String,
    headers: HashMap<String, String>,
}

/// Reason phrase for a status code (the exact text is not significant to the
/// client, which only reads the numeric code and body).
fn reason(status: u16) -> &'static str {
    match status {
        200 => "OK",
        403 => "Forbidden",
        404 => "Not Found",
        500 => "Internal Server Error",
        _ => "Status",
    }
}

/// Reads an HTTP request from the stream, returning the request path and
/// headers.
fn read_request(stream: &mut std::net::TcpStream) -> RecordedRequest {
    let mut buf = Vec::new();
    let mut tmp = [0u8; 1024];
    loop {
        match stream.read(&mut tmp) {
            Ok(0) => break,
            Ok(n) => {
                buf.extend_from_slice(&tmp[..n]);
                if buf.windows(4).any(|w| w == b"\r\n\r\n") {
                    break;
                }
            }
            Err(_) => break,
        }
    }

    let text = String::from_utf8_lossy(&buf);
    let mut lines = text.split("\r\n");
    let request_line = lines.next().unwrap_or("");
    let mut parts = request_line.split_whitespace();
    let _method = parts.next().unwrap_or("");
    let path = parts.next().unwrap_or("").to_string();

    let mut headers = HashMap::new();
    for line in lines {
        if line.is_empty() {
            break;
        }
        if let Some((k, v)) = line.split_once(':') {
            headers.insert(k.trim().to_lowercase(), v.trim().to_string());
        }
    }

    RecordedRequest { path, headers }
}

/// A single-request test server. It accepts one connection, records the
/// request, optionally delays, and responds with a fixed status and body.
struct TestServer {
    url: String,
    last: Arc<Mutex<Option<RecordedRequest>>>,
}

impl TestServer {
    fn spawn(status: u16, body: Vec<u8>, delay: Option<Duration>) -> Self {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let addr = listener.local_addr().unwrap();
        let last: Arc<Mutex<Option<RecordedRequest>>> = Arc::new(Mutex::new(None));
        let last2 = last.clone();

        thread::spawn(move || {
            if let Ok((mut stream, _)) = listener.accept() {
                let req = read_request(&mut stream);
                *last2.lock().unwrap() = Some(req);

                if let Some(d) = delay {
                    thread::sleep(d);
                }

                let head = format!(
                    "HTTP/1.1 {} {}\r\nContent-Length: {}\r\nConnection: close\r\n\r\n",
                    status,
                    reason(status),
                    body.len()
                );
                let _ = stream.write_all(head.as_bytes());
                let _ = stream.write_all(&body);
                let _ = stream.flush();
            }
        });

        TestServer {
            url: format!("http://{addr}"),
            last,
        }
    }

    fn recorded(&self) -> Option<RecordedRequest> {
        self.last.lock().unwrap().clone()
    }
}

fn client_for(url: &str) -> Client {
    Client::with_http_client(ureq::agent(), url)
}

fn download_client_for(url: &str) -> Client {
    Client::with_http_client(ureq::agent(), url).with_download_base_url(url)
}

#[test]
fn test_check_latest_version() {
    struct Case {
        name: &'static str,
        body: &'static str,
        status: u16,
        want_version: &'static str,
        want_err: bool,
    }
    let cases = vec![
        Case {
            name: "valid response",
            body: r#"{"tag_name": "v0.1.0"}"#,
            status: 200,
            want_version: "v0.1.0",
            want_err: false,
        },
        Case {
            name: "valid response with extra fields",
            body: r#"{"tag_name": "v1.2.3", "name": "Release 1.2.3", "draft": false}"#,
            status: 200,
            want_version: "v1.2.3",
            want_err: false,
        },
        Case {
            name: "empty tag_name",
            body: r#"{"tag_name": ""}"#,
            status: 200,
            want_version: "",
            want_err: true,
        },
        Case {
            name: "missing tag_name",
            body: r#"{"name": "Release"}"#,
            status: 200,
            want_version: "",
            want_err: true,
        },
        Case {
            name: "not found",
            body: r#"{"message": "Not Found"}"#,
            status: 404,
            want_version: "",
            want_err: true,
        },
        Case {
            name: "server error",
            body: r#"{"message": "Internal Server Error"}"#,
            status: 500,
            want_version: "",
            want_err: true,
        },
        Case {
            name: "invalid JSON",
            body: "not json",
            status: 200,
            want_version: "",
            want_err: true,
        },
        Case {
            name: "rate limited",
            body: r#"{"message": "API rate limit exceeded"}"#,
            status: 403,
            want_version: "",
            want_err: true,
        },
    ];

    for c in cases {
        let server = TestServer::spawn(c.status, c.body.as_bytes().to_vec(), None);
        let client = client_for(&server.url);
        let result = client.check_latest_version(&Context::background());

        assert_eq!(result.is_err(), c.want_err, "case {}: err mismatch", c.name);
        if let Ok(version) = result {
            assert_eq!(version, c.want_version, "case {}", c.name);
        }

        // Verify request path and headers (checked for the successful cases).
        if !c.want_err {
            let rec = server.recorded().expect("request recorded");
            assert_eq!(
                rec.path, "/repos/Mgrdich/TermChess/releases/latest",
                "case {}: path",
                c.name
            );
            assert_eq!(
                rec.headers.get("accept").map(String::as_str),
                Some("application/vnd.github.v3+json"),
                "case {}: Accept header",
                c.name
            );
            assert!(
                rec.headers.contains_key("user-agent"),
                "case {}: User-Agent header",
                c.name
            );
        }
    }
}

#[test]
fn test_check_latest_version_timeout() {
    // Server sleeps 200ms; context deadline is 50ms.
    let server = TestServer::spawn(
        200,
        br#"{"tag_name": "v0.1.0"}"#.to_vec(),
        Some(Duration::from_millis(200)),
    );
    let client = client_for(&server.url);
    let ctx = Context::with_timeout(Duration::from_millis(50));
    let err = client.check_latest_version(&ctx);
    assert!(err.is_err(), "expected timeout error");
}

#[test]
fn test_check_latest_version_cancellation() {
    let server = TestServer::spawn(
        200,
        br#"{"tag_name": "v0.1.0"}"#.to_vec(),
        Some(Duration::from_millis(200)),
    );
    let client = client_for(&server.url);
    let (ctx, cancel) = Context::with_cancel();
    cancel.cancel(); // cancel immediately
    let err = client.check_latest_version(&ctx);
    assert!(matches!(err, Err(UpdaterError::Cancelled)));
}

#[test]
fn test_download_binary() {
    let expected = b"fake binary data".to_vec();
    let server = TestServer::spawn(200, expected.clone(), None);
    let client = download_client_for(&server.url);

    let data = client
        .download_binary(&Context::background(), "v0.1.0")
        .expect("download");
    assert_eq!(data, expected);

    let rec = server.recorded().expect("request recorded");
    assert!(
        rec.path
            .contains("/releases/download/v0.1.0/termchess-v0.1.0-"),
        "unexpected path: {}",
        rec.path
    );
}

#[test]
fn test_download_binary_error() {
    let server = TestServer::spawn(404, Vec::new(), None);
    let client = download_client_for(&server.url);
    let err = client.download_binary(&Context::background(), "v0.1.0");
    assert!(err.is_err(), "expected error for 404 response");
}

#[test]
fn test_download_checksums() {
    let content = "abc123def456  termchess-v0.1.0-darwin-amd64\n\
                   def789ghi012  termchess-v0.1.0-darwin-arm64";
    let server = TestServer::spawn(200, content.as_bytes().to_vec(), None);
    let client = download_client_for(&server.url);

    let checksums = client
        .download_checksums(&Context::background(), "v0.1.0")
        .expect("download checksums");
    assert_eq!(checksums.len(), 2);
    assert_eq!(checksums["termchess-v0.1.0-darwin-amd64"], "abc123def456");
}

#[test]
fn test_upgrade_already_up_to_date() {
    // Server returns v1.0.0 as latest; current is v1.0.0.
    let server = TestServer::spawn(200, br#"{"tag_name": "v1.0.0"}"#.to_vec(), None);
    let client = client_for(&server.url);
    let err = client.upgrade(&Context::background(), "v1.0.0", "", None);
    assert!(matches!(err, Err(UpdaterError::AlreadyUpToDate)));
}

#[test]
fn test_sentinel_error_messages() {
    assert_eq!(
        UpdaterError::AlreadyUpToDate.to_string(),
        "already up to date"
    );
    assert_eq!(
        UpdaterError::ChecksumMismatch.to_string(),
        "checksum mismatch"
    );
    assert_eq!(
        UpdaterError::PermissionDenied.to_string(),
        "permission denied"
    );
}
