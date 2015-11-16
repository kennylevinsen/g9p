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

	switch mt {
	case Tversion:
		r := &VersionRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rversion:
		r := &VersionResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Tauth:
		r := &AuthRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rauth:
		r := &AuthResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Tattach:
		r := &AttachRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rattach:
		r := &AttachResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rerror:
		r := &ErrorResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Tflush:
		r := &FlushRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rflush:
		r := &FlushResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Twalk:
		r := &WalkRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rwalk:
		r := &WalkResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Topen:
		r := &OpenRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Ropen:
		r := &OpenResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Tcreate:
		r := &CreateRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rcreate:
		r := &CreateResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Tread:
		r := &ReadRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rread:
		r := &ReadResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Twrite:
		r := &WriteRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rwrite:
		r := &WriteResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Tclunk:
		r := &ClunkRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rclunk:
		r := &ClunkResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Tremove:
		r := &RemoveRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rremove:
		r := &RemoveResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Tstat:
		r := &StatRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rstat:
		r := &StatResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Twstat:
		r := &WriteStatRequest{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	case Rwstat:
		r := &WriteStatResponse{}
		if err = r.Decode(limiter); err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, ErrUnknownMessageType
	}
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
