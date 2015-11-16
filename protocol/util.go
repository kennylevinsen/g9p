package protocol

import (
	"encoding/binary"
	"io"
)

// MessageTypeToMessage returns an empty Message based on the provided message
// type.
func MessageTypeToMessage(mt MessageType) (Message, error) {
	switch mt {
	case Tversion:
		return &VersionRequest{}, nil
	case Rversion:
		return &VersionResponse{}, nil
	case Tauth:
		return &AuthRequest{}, nil
	case Rauth:
		return &AuthResponse{}, nil
	case Tattach:
		return &AttachRequest{}, nil
	case Rattach:
		return &AttachResponse{}, nil
	case Tflush:
		return &FlushRequest{}, nil
	case Rflush:
		return &FlushResponse{}, nil
	case Twalk:
		return &WalkRequest{}, nil
	case Rwalk:
		return &WalkResponse{}, nil
	case Topen:
		return &OpenRequest{}, nil
	case Ropen:
		return &OpenResponse{}, nil
	case Tcreate:
		return &CreateRequest{}, nil
	case Rcreate:
		return &CreateResponse{}, nil
	case Tread:
		return &ReadRequest{}, nil
	case Rread:
		return &ReadResponse{}, nil
	case Twrite:
		return &WriteRequest{}, nil
	case Rwrite:
		return &WriteResponse{}, nil
	case Tclunk:
		return &ClunkRequest{}, nil
	case Rclunk:
		return &ClunkResponse{}, nil
	case Tremove:
		return &RemoveRequest{}, nil
	case Rremove:
		return &RemoveRequest{}, nil
	case Tstat:
		return &StatRequest{}, nil
	case Rstat:
		return &StatResponse{}, nil
	case Twstat:
		return &WriteStatRequest{}, nil
	case Rwstat:
		return &WriteStatResponse{}, nil
	case Rerror:
		return &ErrorResponse{}, nil
	default:
		return nil, ErrUnknownMessageType
	}
}

// MessageToMessageType returns the message type of a given message.
func MessageToMessageType(d Message) (MessageType, error) {
	switch d.(type) {
	case *VersionRequest:
		return Tversion, nil
	case *VersionResponse:
		return Rversion, nil
	case *AuthRequest:
		return Tauth, nil
	case *AuthResponse:
		return Rauth, nil
	case *AttachRequest:
		return Tattach, nil
	case *AttachResponse:
		return Rattach, nil
	case *ErrorResponse:
		return Rerror, nil
	case *FlushRequest:
		return Tflush, nil
	case *FlushResponse:
		return Rflush, nil
	case *WalkRequest:
		return Twalk, nil
	case *WalkResponse:
		return Rwalk, nil
	case *OpenRequest:
		return Topen, nil
	case *OpenResponse:
		return Ropen, nil
	case *CreateRequest:
		return Tcreate, nil
	case *CreateResponse:
		return Rcreate, nil
	case *ReadRequest:
		return Tread, nil
	case *ReadResponse:
		return Rread, nil
	case *WriteRequest:
		return Twrite, nil
	case *WriteResponse:
		return Rwrite, nil
	case *ClunkRequest:
		return Tclunk, nil
	case *ClunkResponse:
		return Rclunk, nil
	case *RemoveRequest:
		return Tremove, nil
	case *RemoveResponse:
		return Rremove, nil
	case *StatRequest:
		return Tstat, nil
	case *StatResponse:
		return Rstat, nil
	case *WriteStatRequest:
		return Twstat, nil
	case *WriteStatResponse:
		return Rwstat, nil
	default:
		return Tlast, ErrUnknownMessageType
	}
}

func read(r io.Reader, b []byte) error {
	_, err := io.ReadFull(r, b)
	return err
}

func write(w io.Writer, b []byte) error {
	var (
		written int
		err     error
		l       = len(b)
	)
	for written < l {
		written, err = w.Write(b[written:])
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadByte(r io.Reader) (byte, error) {
	b := make([]byte, 1)
	err := read(r, b)
	if err != nil {
		return 0, err
	}

	return b[0], nil
}

func WriteByte(w io.Writer, b byte) error {
	return write(w, []byte{b})
}

func ReadUint16(r io.Reader) (uint16, error) {
	b := make([]byte, 2)
	err := read(r, b)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(b), nil
}

func WriteUint16(w io.Writer, i uint16) error {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return write(w, b)
}

func ReadUint32(r io.Reader) (uint32, error) {
	b := make([]byte, 4)
	err := read(r, b)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(b), nil
}

func WriteUint32(w io.Writer, i uint32) error {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return write(w, b)
}

func ReadUint64(r io.Reader) (uint64, error) {
	b := make([]byte, 8)
	err := read(r, b)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(b), nil
}

func WriteUint64(w io.Writer, i uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return write(w, b)
}

func ReadTag(r io.Reader) (Tag, error) {
	t, err := ReadUint16(r)
	return Tag(t), err
}

func WriteTag(w io.Writer, t Tag) error {
	return WriteUint16(w, uint16(t))
}

func ReadFid(r io.Reader) (Fid, error) {
	f, err := ReadUint32(r)
	return Fid(f), err
}

func WriteFid(w io.Writer, f Fid) error {
	return WriteUint32(w, uint32(f))
}

func ReadString(r io.Reader) (string, error) {
	l, err := ReadUint16(r)
	if err != nil {
		return "", err
	}

	b := make([]byte, int(l))
	err = read(r, b)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func WriteString(w io.Writer, s string) error {
	err := WriteUint16(w, uint16(len(s)))
	if err != nil {
		return err
	}

	err = write(w, []byte(s))
	if err != nil {
		return err
	}
	return nil
}

func ReadOpenMode(r io.Reader) (OpenMode, error) {
	o, err := ReadByte(r)
	return OpenMode(o), err
}

func WriteOpenMode(w io.Writer, o OpenMode) error {
	return WriteByte(w, byte(o))
}

func ReadQidType(r io.Reader) (QidType, error) {
	o, err := ReadByte(r)
	return QidType(o), err
}

func WriteQidType(w io.Writer, o QidType) error {
	return WriteByte(w, byte(o))
}

func ReadMessageType(r io.Reader) (MessageType, error) {
	mt, err := ReadByte(r)
	return MessageType(mt), err
}

func WriteMessageType(w io.Writer, mt MessageType) error {
	return WriteByte(w, byte(mt))
}

func ReadFileMode(r io.Reader) (FileMode, error) {
	fm, err := ReadUint32(r)
	return FileMode(fm), err
}

func WriteFileMode(w io.Writer, fm FileMode) error {
	return WriteUint32(w, uint32(fm))
}
