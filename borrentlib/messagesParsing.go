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

// ReadMessage reads all other messages which are not handshake
func readMessage(buf io.Reader) (message interface{}, err error) {
	var msg messageBase
	err = binary.Read(buf, binary.BigEndian, &msg)
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
		outMsg := bitfield{}
		toRead := msg.LengthPrefix - 1
		outMsg.bitfield = make([]byte, toRead)
		err = binary.Read(buf, binary.BigEndian, &outMsg.bitfield)
		return outMsg, err
	case pieceMT:
		outMsg := piece{}
		toRead := msg.LengthPrefix - 9
		outMsg.block = make([]byte, toRead)
		err = binary.Read(buf, binary.BigEndian, &outMsg)
		return outMsg, err
	}

	return
}
