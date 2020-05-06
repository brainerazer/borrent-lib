package borrentlib

type handshake struct {
	StrLength uint8 `struct:"uint8,sizeof=Str"`
	Str       []byte
	Reserved  []byte `struct:"[8]byte"`
	InfoHash  []byte `struct:"[20]byte"`
	PeerID    []byte `struct:"[20]byte"`
}
