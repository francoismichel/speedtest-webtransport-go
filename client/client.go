package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"log"
	"ndt-quic-go/ndt"
	"os"
	"time"

	"github.com/lucas-clemente/quic-go/http3"
	"github.com/marten-seemann/webtransport-go"
)

const TEST_DURATION = 10 * time.Second

// DownloadURLPath selects the download subtest.
const DownloadURLPath = "/ndt/vquic/download"

// UploadURLPath selects the upload subtest.
const UploadURLPath = "/ndt/vquic/upload"

type NDTClient struct {
	stats ndt.Stats
}

func (s *NDTClient) endTransferAndSendStats(kind ndt.TransferKind, sess *webtransport.Session) {
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
	addr := flag.String("addr", "127.0.0.1:4443", "address:port to listen on")

	var buf [102400]byte
	var d webtransport.Dialer
	d.RoundTripper = &http3.RoundTripper{}

	d.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	log.Println("Query towards", "https://"+*addr+DownloadURLPath)
	_, session, err := d.Dial(context.Background(), "https://"+*addr+DownloadURLPath, nil)
	if err != nil {
		log.Println("Could not initiate webtransport connection for download", err)
		return
	}

	client := NDTClient{}
	// err is only nil if rsp.StatusCode is a 2xx
	// Handle the connection. Here goes the application logic.
	log.Println("Download")
	ndt.Receive(session, func(n uint64) {
		log.Println("received", n, "bytes")
		client.stats.BytesReceived += n
	}, buf[:], 20*time.Second)
	_, session, err = d.Dial(context.Background(), "https://"+*addr+UploadURLPath, nil)
	if err != nil {
		log.Println("Could not initiate webtransport connection for upload")
		return
	}
	log.Println("Upload")
	ndt.Send(session, buf[:], 20*time.Second)
}
