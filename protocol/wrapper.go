package protocol

import (
	"fmt"
	"io"
)

// Errors
var (
	ErrUnknownMessageType = fmt.Errorf("unknown message type")
)

// Codec is an interface describing an item that can encode itself to a writer,
// decode itself from a reader, and inform how large the encoded form would be
// at the current time.
type Codec interface {
	EncodedLength() int
	Encode(w io.Writer) error
	Decode(r io.Reader) error
}

// Message is an interface like Codec, but also capable of getting/setting the
// message tag. This is merely a convenience feature to save a type assert for
// access to the tag.
type Message interface {
	Codec
	GetTag() Tag
	SetTag(Tag)
}

// DecodeHdr reads 5 bytes and returns the decoded size and message type. It
// may return an error if reading from the Reader fails.
func DecodeHdr(r io.Reader) (uint32, MessageType, error) {
	var (
		size uint32
		mt   MessageType
		err  error
	)

	if size, err = ReadUint32(r); err != nil {
		return 0, 0, err
	}

	if mt, err = ReadMessageType(r); err != nil {
		return 0, 0, err
	}

	return size, mt, nil
}

// Decode decodes an entire message, including header, and returns the message.
// It may return an error if reading from the Reader fails, or if a message
// tries to consume more data than the size of the header indicated, making the
// message invalid.
func Decode(r io.Reader) (Message, error) {
	var (
		size uint32
		mt   MessageType
		err  error
	)
	if size, mt, err = DecodeHdr(r); err != nil {
		return nil, err
	}

	// This LimitedReader is not a necessity, but used as an error checker.
	limiter := &io.LimitedReader{R: r, N: int64(size) - HeaderSize}

	m, err := MessageTypeToMessage(mt)
	if err != nil {
		return nil, err
	}
	if err = m.Decode(limiter); err != nil {
		return nil, err
	}
	return m, nil
}

// Encode write a header and message to the provided writer. It returns an
// error if writing failed.
func Encode(w io.Writer, d Message) error {
	var err error
	var mt MessageType
	if mt, err = MessageToMessageType(d); err != nil {
		return err
	}

	size := uint32(d.EncodedLength() + HeaderSize)
	if err = WriteUint32(w, size); err != nil {
		return err
	}

	if err = WriteMessageType(w, mt); err != nil {
		return err
	}

	if err = d.Encode(w); err != nil {
		return err
	}
	return nil
}
