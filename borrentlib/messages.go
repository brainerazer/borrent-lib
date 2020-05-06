package borrentlib

type handshake struct {
	StrLength uint8 `struct:"uint8,sizeof=Str"`
	Str       []byte
	Reserved  []byte `struct:"[8]byte"`
	InfoHash  []byte `struct:"[20]byte"`
	PeerID    []byte `struct:"[20]byte"`
}

type messageType byte

const (
	chokeMT messageType = iota
	unchokeMT
	interestedMT
	notInterestedMT
	haveMT
	bitfieldMT
	requestMT
	pieceMT
	cancelMT
	portMT
)

type messageBase struct {
	LengthPrefix uint32
	MessageID    messageType
}

type keepAlive struct {
	messageBase // No MessageID though (LengthPrefix = 0)!
}

type choke struct {
	messageBase
}

type unchoke struct {
}

type interested struct {
}

type notInterested struct {
}

type have struct {
	pieceIndex uint32
}

type bitfield struct {
	bitfield []byte
}

type request struct {
	index  uint32
	begin  uint32
	length uint32
}

type piece struct {
	index uint32
	begin uint32
	block []byte
}

type cancel struct {
	index  uint32
	begin  uint32
	length uint32
}

type port struct {
	listenPort uint16
}
