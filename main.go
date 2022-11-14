package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"ndt-quic-go/ndt"

	"github.com/lucas-clemente/quic-go/http3"
	"github.com/marten-seemann/webtransport-go"
)

const TEST_DURATION = 10*time.Second

// DownloadURLPath selects the download subtest.
const DownloadURLPath = "/ndt/vquic/download"

// UploadURLPath selects the upload subtest.
const UploadURLPath = "/ndt/vquic/upload"

type NDTServer struct {
	stats ndt.Stats
}

func (s *NDTServer) endTransferAndSendStats(kind ndt.TransferKind, sess *webtransport.Session) {
	s.stats.ElapsedTime = time.Now().Sub(s.stats.StartTime)
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

		str.Write([]byte(""))
	}
}

func main() {
	htmlDir := "." 

	www := flag.String("www", ".", "HTTP root directory")
	certFile := flag.String("cert", "cert.pem", "path to the certificate")
	keyFile := flag.String("key", "key.pem", "path to the private key")
	addr := flag.String("addr", ":4443", "address:port to listen on")

	// The ndt7 listener serving up NDT7 tests, likely on standard ports.
	ndt7Mux := http.NewServeMux()
	ndt7Mux.Handle(*www, http.FileServer(http.Dir(htmlDir)))
	// create a new webtransport.Server, listening on (UDP) port 443
	server := NDTServer{
		stats: ndt.Stats {
			BytesReceived: 0,
			StartTime: time.Now(),
			ElapsedTime: 0,
		},
	}


	h3Server := http3.Server{
		Addr: *addr,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
		Handler: ndt7Mux,
	}
	handler := ndt.NDT7Handler{
		Server: &webtransport.Server{
			H3: h3Server,
		},
		ReceiveCallback: func(n uint64) {
			log.Println("received", n, "bytes")
			server.stats.BytesReceived += n
		},
		TransferEndCallback: server.endTransferAndSendStats,
		TestDuration: TEST_DURATION,
	}
	
	ndt7Mux.HandleFunc(DownloadURLPath, handler.UpgradeAndSend)
	ndt7Mux.HandleFunc(UploadURLPath, handler.UpgradeAndReceive)


	handler.Server.ListenAndServeTLS(*certFile, *keyFile)
}