package protocol

import (
	"encoding/binary"
	"io"
	"math"
)

type Protocol interface {
	Size() int
	MarshalBinary() ([]byte, error)
	UnmarshalBinary([]byte) error
	Validate() error
}

func ReadUint8(r io.Reader) (uint8, error) {
	var buf [1]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func ReadUint16(r io.Reader) (uint16, error) {
	var buf [2]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf[:]), nil
}

func ReadUint32(r io.Reader) (uint32, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf[:]), nil
}

func ReadUint64(r io.Reader) (uint64, error) {
	var buf [8]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf[:]), nil
}

func ReadFloat32(r io.Reader) (float32, error) {
	bits, err := ReadUint32(r)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(bits), nil
}

func ReadFloat64(r io.Reader) (float64, error) {
	bits, err := ReadUint64(r)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

func WriteUint8(w io.Writer, v uint8) error {
	_, err := w.Write([]byte{v})
	return err
}

func WriteUint16(w io.Writer, v uint16) error {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

func WriteUint32(w io.Writer, v uint32) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

func WriteUint64(w io.Writer, v uint64) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

func WriteFloat32(w io.Writer, v float32) error {
	return WriteUint32(w, math.Float32bits(v))
}

func WriteFloat64(w io.Writer, v float64) error {
	return WriteUint64(w, math.Float64bits(v))
}
