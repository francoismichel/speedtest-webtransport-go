package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"ndt-quic-go/ndt"

	"github.com/lucas-clemente/quic-go/http3"
	"github.com/lucas-clemente/quic-go/interop/utils"
	"github.com/marten-seemann/webtransport-go"
)

const TEST_DURATION = 10 * time.Second

// DownloadURLPath selects the download subtest.
const DownloadURLPath = "/ndt/vquic/download"

// UploadURLPath selects the upload subtest.
const UploadURLPath = "/ndt/vquic/upload"

type NDTServer struct {
	stats ndt.Stats
}

func (s *NDTServer) endTransferAndSendStats(kind ndt.TransferKind, sess *webtransport.Session) {
	s.stats.ElapsedTime = time.Duration(time.Now().Sub(s.stats.StartTime).Microseconds())
	if kind == ndt.TransferReceive {
		str, err := sess.OpenUniStream()
		if err != nil {
			log.Println("Could not open stream for sending statistics")
			return
		}
		defer str.Close()
		// TODO
		encoder := json.NewEncoder(str)
		encoder.Encode(s.stats)

		stdoutEncoder := json.NewEncoder(os.Stdout)
		stdoutEncoder.Encode(s.stats)
		s.stats.BytesReceived = 0
		s.stats.StartTime = time.UnixMilli(0)
	}
}

func main() {
	htmlDir := "."

	www := flag.String("www", "www", "HTTP root directory")
	certFile := flag.String("cert", "cert.pem", "path to the certificate")
	keyFile := flag.String("key", "key.pem", "path to the private key")
	hostname := flag.String("hostname", "localhost", "server hostname")
	addr := flag.String("addr", ":4443", "address:port to listen on")
	flag.Parse()

	keyLog, err := utils.GetSSLKeyLog()
	if err != nil {
		fmt.Printf("Could not create key log: %s\n", err.Error())
		os.Exit(1)
	}
	if keyLog != nil {
		defer keyLog.Close()
	}

	// The ndt7 listener serving up NDT7 tests, likely on standard ports.
	ndt7Mux := http.NewServeMux()
	ndt7Mux.Handle(*www, http.FileServer(http.Dir(htmlDir)))
	// create a new webtransport.Server, listening on (UDP) port 443
	server := NDTServer{
		stats: ndt.Stats{
			BytesReceived: 0,
			StartTime:     time.UnixMilli(0),
			ElapsedTime:   0,
		},
	}
	certs := make([]tls.Certificate, 1)
	certs[0], _ = tls.LoadX509KeyPair(*certFile, *keyFile)
	h3Server := http3.Server{
		Addr: *addr,
		TLSConfig: &tls.Config{
			Certificates: certs,
			ServerName:   *hostname,
			KeyLogWriter: keyLog,
		},
		Handler: ndt7Mux,
	}
	handler := ndt.NDT7Handler{
		Server: &webtransport.Server{
			H3: h3Server,
		},
		ReceiveCallback: func(n uint64) {
			if server.stats.StartTime == time.UnixMilli(0) {
				server.stats.StartTime = time.Now()
			}
			server.stats.BytesReceived += n
			//TODO(mp): Send that to the client somehow
		},
		TransferEndCallback: server.endTransferAndSendStats,
		TestDuration:        TEST_DURATION,
	}

	ndt7Mux.HandleFunc(DownloadURLPath, handler.UpgradeAndSend)
	ndt7Mux.HandleFunc(UploadURLPath, handler.UpgradeAndReceive)

	handler.Server.ListenAndServe()
}
