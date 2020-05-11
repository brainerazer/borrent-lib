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
func ReadMessage(buf io.Reader) (message torrentMessage, err error) {
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
		return Unchoke{}, nil
	case interestedMT:
		return Interested{}, nil
	case notInterestedMT:
		return notInterested{}, nil
	case haveMT:
		var outMsg have
		err = binary.Read(buf, binary.BigEndian, &outMsg)
		return outMsg, err
	case requestMT:
		var outMsg Request
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
	err = message.WriteTo(buf)
	return err
}

func (msg keepAlive) WriteTo(w io.Writer) error {
	_, err := w.Write([]byte("\x00\x00\x00\x00"))
	return err
}

func (msg choke) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{1, chokeMT})
	return err
}

func (msg Unchoke) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{1, unchokeMT})
	return err
}

func (msg Interested) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{1, interestedMT})
	return err
}

func (msg notInterested) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{1, notInterestedMT})
	return err
}

func (msg have) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{5, haveMT})
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, msg)
	return err
}

func (msg Request) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{13, requestMT})
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, msg)
	return err
}

func (msg cancel) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{13, cancelMT})
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, msg)
	return err
}

func (msg port) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{3, portMT})
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, msg)
	return err
}

func (msg Bitfield) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{1 + uint32(len(msg.Bitfield)), bitfieldMT})
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, msg.Bitfield)
	return err
}

func (msg Piece) WriteTo(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, messageBase{9 + uint32(len(msg.Block)), pieceMT})
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, msg.Index)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, msg.Begin)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, msg.Block)
	if err != nil {
		return err
	}
	return err
}
