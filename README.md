# speedtest-webtransport-go
Simple speedtest server and client using webtransport over HTTP/3

The server handles webtransport sessions on these two endpoints:
- `/ndt/vquic/download`: sends data on a single stream during 10 seconds, receives final stats on another one
- `/ndt/vquic/upload`: waits for data on a single stream for 10 seconds, sends final stats on another one

A successful execution of a test on an endpoint produces a result in the following format:
```
{
    "TransferKind": 0, // 0 => Client download, 1 => Client upload
    "BytesReceived": 914620416,
    "StartTime": "2022-12-13T13:20:19.611522162+01:00", // RFC 3339 format
    "ElapsedTime": 10743504  // µs
}
```

Each test result is printed on stdout.

## Web client

`www/index.html` contains a simple webpage triggering the download and upload tests.
It displays the obtained bandwidth. The `?target=` search parameter allows setting the speedtest server hostname
when it differs from the current hostname.