// Package packet implements marshalling and unmarshalling of TFTP packets, as
// defined in rfc1350.
//
// Unfortunately, the rfc doesn't say anything about byte order. However, it is
// possible to detect the byte order of an incoming packet. This package
// reports the byte order of an incoming packet to the user. Most likely, the
// user will want to transmit response packets in the same byte order as the
// original packet.
package packet

import (
	"bytes"
	"encoding/binary"
	"io"
)

const (
	// opcodes
	RRQ = 1
	WRQ = 2
	DATA = 3
	ACK = 4
	ERROR = 5
)

const (
	// error codes
	ErrNotDefined = 0
	ErrFileNotFound = 1
	ErrAccess = 2
	ErrDiskFull = 3
	ErrIllegalOp = 4
	ErrUnknownTID = 5
	ErrFileExists = 6
	ErrBadUser = 7
)

type Rq struct {
	Filename string
	Mode string
}

type Rrq Rq
type Wrq Rq

type Data struct {
	BlockNum uint16
	Data []byte
}

type Ack struct {
	BlockNum uint16
}

type Error struct {
	ErrorCode uint16
	ErrorMsg string
}

// A TFTP packet. All of the methods of this interface are private. The types
// *Rrq, *Wrq, *Data, *Ack, and *Error implement the Packet interface.
type Packet interface{
	readFrom(r io.Reader, order binary.ByteOrder) error
}

// Read a TFTP packet from r, returning the packet, its byte order, and an
// error. If the error is non-nil, then the packet and byte order are invalid.
// The byte order is inferred from the opcode of the packet.
func ReadPacket(r io.Reader) (Packet, binary.ByteOrder, error) {

	// Rfc1350 doesn't say anything about byte order, but we can detect it, since
	// only opcodes 1 through 5 are valid. We try reading it in as little endian,
	// and if we get something invalid, we assume we picked the wrong byte order.
	var order binary.ByteOrder
	order = binary.LittleEndian
	var Opcode uint16
	err := binary.Read(r, order, &Opcode)
	if err != nil {
		return nil, nil, err
	}
	if Opcode > ERROR {
		// Wrong endianness; convert it and change the order for later.
		Opcode = Opcode >> 8
		order = binary.BigEndian
	}

	var ret Packet

	switch Opcode {
	case RRQ:
		ret = &Rrq{}
	case WRQ:
		ret = &Wrq{}
	case DATA:
		ret = &Data{}
	case ACK:
		ret = &Ack{}
	case ERROR:
		ret = &Error{}
	}

	err = ret.readFrom(r, order)
	return ret, order, err
}


func (rq *Rq) readFrom(r io.Reader, order binary.ByteOrder) error {
	err := readString(r, &rq.Filename)
	if err != nil {
		return err
	}
	err = readString(r, &rq.Mode)
	return err
}

func (req *Rrq) readFrom(r io.Reader, order binary.ByteOrder) error {
	return (*Rq)(req).readFrom(r, order)
}

func (req *Wrq) readFrom(r io.Reader, order binary.ByteOrder) error {
	return (*Rq)(req).readFrom(r, order)
}

func (d *Data) readFrom(r io.Reader, order binary.ByteOrder) error {
	err := binary.Read(r, order, &d.BlockNum)
	if err != nil {
		return err
	}
	d.Data = make([]byte, 512)
	_, err = r.Read(d.Data)
	return err
}

func (a *Ack) readFrom(r io.Reader, order binary.ByteOrder) error {
	return binary.Read(r, order, &a.BlockNum)
}

func (e *Error) readFrom(r io.Reader, order binary.ByteOrder) error {
	err := binary.Read(r, order, &e.ErrorCode)
	if err != nil {
		return err
	}
	err = readString(r, &e.ErrorMsg)
	return err
}

func readString(r io.Reader, s *string) error {
	buf := bytes.Buffer{}
	ch := []byte{0}
	_, err := r.Read(ch)
	if err != nil {
		return err
	}
	for ch[0] != 0 {
		buf.Write(ch)
		_, err = r.Read(ch)
		if err != nil {
			return err
		}
	}
	*s = buf.String()
	return nil
}

func writeString(s string, w io.Writer) (n int, err error) {
	nS, err := w.Write([]byte(s))
	if err != nil {
		return nS, err
	}
	n0, err := w.Write([]byte{0})
	return nS + n0, err
}
