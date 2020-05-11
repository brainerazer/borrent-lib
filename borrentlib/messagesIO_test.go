package borrentlib

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func helperReadAll(t *testing.T, r io.Reader) []byte {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

var readWriteTestData = []struct {
	testName string
	message  torrentMessage
	rawBytes []byte
	wantErr  bool
}{
	{
		"Wireshark sample no 1 - Bitfield",
		Bitfield{Bitfield: []byte("\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
			"\xff\xff\xff\xff\xff\xff\xff\xfe")},
		[]byte("\x00\x00\x00\x19\x05\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
			"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xfe"),
		false,
	},
	{
		"Wireshark sample no 2 - Have",
		have{PieceIndex: 0x000000a0},
		[]byte("\x00\x00\x00\x05\x04\x00\x00\x00\xa0"),
		false,
	},
	{
		"Wireshark sample no 2 - Interested",
		Interested{},
		[]byte("\x00\x00\x00\x01\x02"),
		false,
	},
	{
		"Wireshark sample no 2 - Unchoke",
		Unchoke{},
		[]byte("\x00\x00\x00\x01\x01"),
		false,
	},
	{
		"Wireshark sample no 2 - Request",
		Request{Index: 0x00000048, Begin: 0x00000000, Length: 0x00004000},
		[]byte("\x00\x00\x00\x0d\x06\x00\x00\x00\x48\x00\x00\x00\x00\x00\x00\x40\x00"),
		false,
	},
	{
		"Wireshark own - choke",
		choke{},
		[]byte("\x00\x00\x00\x01\x00"),
		false,
	},
	{
		"Wireshark NOT CAPTURE (couldn't find  one) - keepalive",
		keepAlive{},
		[]byte("\x00\x00\x00\x00"),
		false,
	},
}

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
	for _, tt := range readWriteTestData {
		t.Run(tt.testName, func(t *testing.T) {
			gotMessage, err := ReadMessage(bytes.NewReader(tt.rawBytes))
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMessage, tt.message) {
				t.Errorf("ReadMessage() = %v, want %v", gotMessage, tt.message)
			}
		})
	}
}

func Test_readMessage_largepiece(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name        string
		args        args
		wantMessage interface{}
		wantErr     bool
	}{
		{
			"Wireshark own - piece",
			args{
				"piece_message.bin",
			},
			Piece{Index: 0x00000ee2, Begin: 0x00018000,
				Block: helperReadAll(t, helperLoadFile(t, "piece_data.bin"))},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := helperLoadFile(t, tt.args.filename)
			gotMessage, err := ReadMessage(file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantMessage, gotMessage); diff != "" {
				t.Errorf("ReadMessage()  mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriteMessage(t *testing.T) {
	for _, tt := range readWriteTestData {
		t.Run(tt.testName, func(t *testing.T) {
			buf := &bytes.Buffer{}
			if err := WriteMessage(buf, tt.message); (err != nil) != tt.wantErr {
				t.Errorf("WriteMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBuf := buf.Bytes(); !bytes.Equal(gotBuf, tt.rawBytes) {
				t.Errorf("WriteMessage() = %v, want %v", gotBuf, tt.rawBytes)
			}
		})
	}
}

func Test_WriteMessage_largepiece(t *testing.T) {
	type args struct {
		message Piece
	}
	tests := []struct {
		name         string
		args         args
		wantFilename string
		wantErr      bool
	}{
		{
			"Wireshark own - piece",
			args{
				Piece{Index: 0x00000ee2, Begin: 0x00018000,
					Block: helperReadAll(t, helperLoadFile(t, "piece_data.bin"))},
			},
			"piece_message.bin",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := helperReadAll(t, helperLoadFile(t, tt.wantFilename))
			buf := &bytes.Buffer{}
			err := WriteMessage(buf, tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(file, buf.Bytes()); diff != "" {
				t.Errorf("WriteMessage()  mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
