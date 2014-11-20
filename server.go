package tftp

import (
	"github.com/zenhack/go.tftp/packet"
)


type GetReq struct {
	in <-chan packet.Packet
	out chan<- packet.Packet
	data []byte
}

type PutReq struct {
	in <-chan packet.Packet
	out chan<- packet.Packet
	data []byte
}

func (r *GetReq) Write(p []byte) (n int, err error) {
}

func (r *PutReq) Read(p []byte) (n int, err error) {
	goal := len(p)
	soFar := 0
	for soFar < goal {
		if len(r.data) == 0 {
			pkt := <-r.in
			dataPkt, ok := pkt.(*packet.Data)
			if !ok {
				panic(4)
			}
			r.data = dataPkt.Data
		}
	}
}

func (r *GetReq) RespondError(code int, msg string) {
}

func (r *PutReq) RespondError(code int, msg string) {
}

func handleClient(in <-chan packet.Packet, out chan<- packet.Packet) {
	pkt := <-in
	switch p := pkt.(type) {
	case *packet.Rrq:
	case *packet.Wrq:
		handleWrite(p)
	default:
	}
}
