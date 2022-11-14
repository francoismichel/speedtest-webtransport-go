# speedtest-webtransport-go
Simple speedtest server and client using webtransport over HTTP/3

The server handles webtransport sessions on these two endpoints:
- `/ndt/vquic/download`: sends data on a single stream during 10 seconds
- `/ndt/vquic/upload`: waits for data on a single stream for 10 seconds

## TODO
JSON stats reporting from the client for the download and from the server for the upload. 
