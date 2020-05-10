package borrentlib

import (
	"encoding/binary"
	"io"

	"gopkg.in/restruct.v1"
)

// ReadHandshake reads a handshake message
func readHandshake(buf io.Reader) (hs handshake, err error) {
	decoded := make([]byte, 68)

	_, err = io.ReadFull(buf, decoded)
	if err != nil {
		return
	}

	err = restruct.Unpack(decoded, binary.LittleEndian, &hs)
	return hs, nil
}

func writeHandshake(buf io.Writer, hs *handshake) (err error) {
	encoded, err := restruct.Pack(binary.LittleEndian, hs)
	if err != nil {
		return err
	}

	_, err = buf.Write(encoded)
	if err != nil {
		return err
	}
	return nil
}

func readHeader(buf io.Reader) (header messageBase, err error) {
	err = binary.Read(buf, binary.BigEndian, &header.LengthPrefix)
	if err != nil {
		return
	}

	if header.LengthPrefix == 0 { // seems like keep-alive - not id to read
		return header, nil
	}

	err = binary.Read(buf, binary.BigEndian, &header.MessageID)
	if err != nil {
		return
	}

	return header, nil
}

// ReadMessage reads all other messages which are not handshake
func readMessage(buf io.Reader) (message torrentMessage, err error) {
	msg, err := readHeader(buf)
	if err != nil {
		return
	}

	if msg.LengthPrefix == 0 { // keep-alive
		return keepAlive{}, nil
	}

	switch msg.MessageID {

	// All the fixed-size first, as they are easier
	case chokeMT:
		return choke{}, nil
	case unchokeMT:
		return unchoke{}, nil
	case interestedMT:
		return interested{}, nil
	case notInterestedMT:
		return notInterested{}, nil
	case haveMT:
		var outMsg have
		err = binary.Read(buf, binary.BigEndian, &outMsg)
		return outMsg, err
	case requestMT:
		var outMsg request
		err = binary.Read(buf, binary.BigEndian, &outMsg)
		return outMsg, err
	case cancelMT:
		var outMsg cancel
		err = binary.Read(buf, binary.BigEndian, &outMsg)
		return outMsg, err
	case portMT:
		var outMsg port
		err = binary.Read(buf, binary.BigEndian, &outMsg)
		return outMsg, err

	// Now to variable-size
	case bitfieldMT:
		outMsg := Bitfield{}
		toRead := msg.LengthPrefix - 1
		outMsg.Bitfield = make([]byte, toRead)
		err = binary.Read(buf, binary.BigEndian, &outMsg.Bitfield)
		return outMsg, err
	case pieceMT:
		outMsg := Piece{}
		toRead := msg.LengthPrefix - 9
		outMsg.Block = make([]byte, toRead)
		err = binary.Read(buf, binary.BigEndian, &outMsg.Index)
		if err != nil {
			return nil, err
		}
		err = binary.Read(buf, binary.BigEndian, &outMsg.Begin)
		if err != nil {
			return nil, err
		}
		err = binary.Read(buf, binary.BigEndian, &outMsg.Block)
		return outMsg, err
	}

	return
}

// WriteMessage ...
func WriteMessage(buf io.Writer, message torrentMessage) (err error) {
	bytes, err := message.toBytes()
	if err != nil {
		return
	}
	_, err = buf.Write(bytes)
	return
}

func (msg keepAlive) toBytes() ([]byte, error) {
	return []byte("\x00\x00\x00\x00"), nil
}

func (msg choke) toBytes() ([]byte, error) {
	panic("Not implemented!")
}

func (msg unchoke) toBytes() ([]byte, error) {
	panic("Not implemented!")
}

func (msg interested) toBytes() ([]byte, error) {
	panic("Not implemented!")
}

func (msg notInterested) toBytes() ([]byte, error) {
	panic("Not implemented!")
}

func (msg have) toBytes() ([]byte, error) {
	panic("Not implemented!")
}

func (msg request) toBytes() ([]byte, error) {
	panic("Not implemented!")
}

func (msg cancel) toBytes() ([]byte, error) {
	panic("Not implemented!")
}

func (msg port) toBytes() ([]byte, error) {
	panic("Not implemented!")
}

func (msg Bitfield) toBytes() ([]byte, error) {
	panic("Not implemented!")
}

func (msg Piece) toBytes() ([]byte, error) {
	panic("Not implemented!")
}
