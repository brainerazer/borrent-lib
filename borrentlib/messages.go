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

// No MessageID (LengthPrefix = 0)!
type keepAlive struct {
}

type choke struct {
}

type unchoke struct {
}

type interested struct {
}

type notInterested struct {
}

type have struct {
	PieceIndex uint32
}

type Bitfield struct {
	Bitfield []byte
}

type request struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

type Piece struct {
	Index uint32
	Begin uint32
	Block []byte
}

type cancel struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

type port struct {
	ListenPort uint16
}
