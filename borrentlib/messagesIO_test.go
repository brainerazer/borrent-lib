package borrentlib

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func Test_readHandshake(t *testing.T) {
	type args struct {
		buf io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantHs  handshake
		wantErr bool
	}{
		{
			"Wireshark sample no 1",
			args{
				bytes.NewReader([]byte("\x13\x42\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74" +
					"\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x00\x00\x00\x01\x64\xfe\x7e" +
					"\xf1\x10\x5c\x57\x76\x41\x70\xed\xf6\x03\xc4\x39\xd6\x42\x14\xf1" +
					"\x2d\x41\x5a\x32\x32\x30\x32\x2d\x38\x72\x44\x4e\x59\x41\x67\x6e" +
					"\x36\x51\x66\x39")),
			},
			handshake{
				StrLength: 19, Str: []byte("BitTorrent protocol"), Reserved: make([]byte, 8),
				InfoHash: []byte("\x01\x64\xfe\x7e\xf1\x10\x5c\x57\x76\x41\x70\xed\xf6\x03\xc4\x39" +
					"\xd6\x42\x14\xf1"),
				PeerID: []byte("\x2d\x41\x5a\x32\x32\x30\x32\x2d\x38\x72\x44\x4e\x59\x41\x67\x6e" +
					"\x36\x51\x66\x39"),
			},
			false,
		},
		{
			"Wireshark sample no 1 - extbytes",
			args{
				bytes.NewReader([]byte("\x13\x42\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74" +
					"\x6f\x63\x6f\x6c\x65\x78\x00\x00\x00\x00\x00\x00\x01\x64\xfe\x7e" +
					"\xf1\x10\x5c\x57\x76\x41\x70\xed\xf6\x03\xc4\x39\xd6\x42\x14\xf1" +
					"\x65\x78\x62\x63\x00\x38\x31\x7b\x01\x75\x33\xf4\x1b\x14\x11\xa8" +
					"\xab\x28\xbb\x54")),
			},
			handshake{
				StrLength: 19, Str: []byte("BitTorrent protocol"),
				Reserved: []byte("\x65\x78\x00\x00\x00\x00\x00\x00"),
				InfoHash: []byte("\x01\x64\xfe\x7e\xf1\x10\x5c\x57\x76\x41\x70\xed\xf6\x03\xc4\x39" +
					"\xd6\x42\x14\xf1"),
				PeerID: []byte("\x65\x78\x62\x63\x00\x38\x31\x7b\x01\x75\x33\xf4\x1b\x14\x11\xa8" +
					"\xab\x28\xbb\x54"),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHs, err := readHandshake(tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("readHandshake() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotHs, tt.wantHs) {
				t.Errorf("readHandshake() = %v, want %v", gotHs, tt.wantHs)
			}
		})
	}
}

func Test_writeHandshake(t *testing.T) {
	type args struct {
		hs *handshake
	}
	tests := []struct {
		name    string
		args    args
		wantBuf string
		wantErr bool
	}{
		{
			"Wireshark sample no 1",
			args{
				&handshake{
					StrLength: 19, Str: []byte("BitTorrent protocol"), Reserved: make([]byte, 8),
					InfoHash: []byte("\x01\x64\xfe\x7e\xf1\x10\x5c\x57\x76\x41\x70\xed\xf6\x03\xc4\x39" +
						"\xd6\x42\x14\xf1"),
					PeerID: []byte("\x2d\x41\x5a\x32\x32\x30\x32\x2d\x38\x72\x44\x4e\x59\x41\x67\x6e" +
						"\x36\x51\x66\x39"),
				},
			},
			"\x13\x42\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74" +
				"\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x00\x00\x00\x01\x64\xfe\x7e" +
				"\xf1\x10\x5c\x57\x76\x41\x70\xed\xf6\x03\xc4\x39\xd6\x42\x14\xf1" +
				"\x2d\x41\x5a\x32\x32\x30\x32\x2d\x38\x72\x44\x4e\x59\x41\x67\x6e" +
				"\x36\x51\x66\x39",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			if err := writeHandshake(buf, tt.args.hs); (err != nil) != tt.wantErr {
				t.Errorf("writeHandshake() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBuf := buf.String(); gotBuf != tt.wantBuf {
				t.Errorf("writeHandshake() = %v, want %v", gotBuf, tt.wantBuf)
			}
		})
	}
}

func Test_readMessage(t *testing.T) {
	type args struct {
		buf io.Reader
	}
	tests := []struct {
		name        string
		args        args
		wantMessage interface{}
		wantErr     bool
	}{
		{
			"Wireshark sample no 1 - Bitfield",
			args{
				bytes.NewReader([]byte("\x00\x00\x00\x19\x05\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
					"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xfe")),
			},
			Bitfield{Bitfield: []byte("\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
				"\xff\xff\xff\xff\xff\xff\xff\xfe")},
			false,
		},
		{
			"Wireshark sample no 2 - Have",
			args{
				bytes.NewReader([]byte("\x00\x00\x00\x05\x04\x00\x00\x00\xa0")),
			},
			have{PieceIndex: 0x000000a0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMessage, err := readMessage(tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("readMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMessage, tt.wantMessage) {
				t.Errorf("readMessage() = %v, want %v", gotMessage, tt.wantMessage)
			}
		})
	}
}
