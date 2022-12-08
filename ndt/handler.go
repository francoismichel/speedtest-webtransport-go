package ndt

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/marten-seemann/webtransport-go"
)

type TransferKind int

const (
	TransferSend TransferKind = iota
	TransferReceive
)

type NDT7Handler struct {
	Server              *webtransport.Server
	ReceiveCallback     func(uint64)
	TransferEndCallback func(TransferKind, *webtransport.Session)
	TestDuration        time.Duration
}

func (h *NDT7Handler) UpgradeAndSend(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 102400)
	n, err := rand.Read(buf)
	if err != nil || n != 102400 {
		log.Println("Could not generate random payload.")
		return
	}
	log.Println("CONNECT AND SEND")
	session, err := h.Server.Upgrade(w, r)
	if err != nil {
		log.Println("Could not create WebTransport session for sending.")
		return
	}
	h.TransferEndCallback(TransferSend, session)
	defer session.Close()
}

func (h *NDT7Handler) UpgradeAndReceive(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 102400)
	session, err := h.Server.Upgrade(w, r)
	if err != nil {
		log.Println("Could not create WebTransport session for receiving.")
		return
	}
	defer session.Close()
	Receive(session, h.ReceiveCallback, buf[:], h.TestDuration)
	h.TransferEndCallback(TransferReceive, session)
}

func Send(session *webtransport.Session, buf []byte, testDuration time.Duration) {
	str, err := session.OpenUniStream()
	if err != nil {
		log.Println("Could not open stream for sending.")
		return
	}
	defer str.Close()

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Println("Sending done.")
			return
		default:
			_, err := str.Write(buf)
			if err != nil {
				log.Println("Could not write bytes on stream for sending:", err)
				return
			}
		}
	}
}

func Receive(session *webtransport.Session, receiveCallback func(uint64), buf []byte, testDuration time.Duration) {
	deadline := time.Now().Add(testDuration)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	str, err := session.AcceptUniStream(ctx)
	if err != nil {
		log.Println("Could not get stream from the peer.")
		return
	}
	err = str.SetReadDeadline(deadline)
	if err != nil {
		log.Println("Could not set the read deadline on the stream for receiving.")
	}
	for {
		select {
		case <-ctx.Done():
			log.Println("Receiving done.")
			return
		default:
			n, err := str.Read(buf)
			if n > 0 && receiveCallback != nil {
				receiveCallback(uint64(n))
			}
			if err != nil {
				if errors.Unwrap(err) == os.ErrDeadlineExceeded {
					log.Println("Receiving done.")
					return
				}
			}
		}
	}
}
