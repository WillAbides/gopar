package par2

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"reflect"
)

type packetHeader struct {
	Magic         [8]byte
	Length        uint64
	Hash          [16]byte
	RecoverySetID [16]byte
	Type          [16]byte
}

var expectedMagic = [8]byte{'P', 'A', 'R', '2', '\x00', 'P', 'K', 'T'}

func sizeOfPacketHeader() uint64 {
	return uint64(reflect.TypeOf(packetHeader{}).Size())
}

func readPacketHeader(buf *bytes.Buffer) (packetHeader, error) {
	var h packetHeader
	err := binary.Read(buf, binary.LittleEndian, &h)
	if err != nil {
		return packetHeader{}, err
	}

	if h.Magic != expectedMagic {
		return packetHeader{}, errors.New("unexpected magic string")
	}

	if h.Length < sizeOfPacketHeader() || h.Length%4 != 0 {
		return packetHeader{}, errors.New("invalid length")
	}

	return h, nil
}

type recoverySetID [16]byte
type packetType [16]byte

func readNextPacket(buf *bytes.Buffer) (recoverySetID, packetType, []byte, error) {
	h, err := readPacketHeader(buf)
	if err != nil {
		return [16]byte{}, packetType{}, nil, err
	}

	// TODO: Handle overflow.
	bodyLength := int(h.Length - sizeOfPacketHeader())
	body := buf.Next(bodyLength)
	if len(body) != bodyLength {
		return [16]byte{}, packetType{}, nil, errors.New("could not read body")
	}

	var hashInput []byte
	hashInput = append(hashInput, h.RecoverySetID[:]...)
	hashInput = append(hashInput, h.Type[:]...)
	hashInput = append(hashInput, body...)
	if md5.Sum(hashInput) != h.Hash {
		return [16]byte{}, packetType{}, nil, errors.New("hash mismatch")
	}

	bodyCopy := make([]byte, len(body))
	copy(bodyCopy, body)
	return h.RecoverySetID, h.Type, bodyCopy, nil
}
