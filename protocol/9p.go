/*
Package protocol implements the 9P2000 protocol.

This module contains definitions of the protocol messages. In some cases, the
struct skips fields that are redundant in a Go context, such as a fields that
just tell the length of variable length arrays, which is already handled by
slice headers.
*/
package protocol

import "io"

// MessageType is the type of the contained message.
type MessageType byte

// IsRequest checks if a message type is a request or a response.
func (m MessageType) IsRequest() bool {
	return byte(m)%2 == 0
}

//
// Types that are part of messages below.
//

// Tag is a unique identifier for a request. It is echoed by the response. It
// is the responsibility of the client to ensure that it is unique among all
// current requests.
type Tag uint16

// Fid, or "file identifier", is quite similar in concept to a file descriptor,
// and is used to keep track of a file and its potential opening mode. The
// client is responsible for providing a unique Fid to use. The Fid is passed
// to all later requests to inform the server what file the manipulation should
// occur on. Multiple Fids can be open on a connection. Fids are local to the
// connection, and can be reused after Remove or Clunk.
type Fid uint32

// OpenMode, as the name implies, represents the opening mode of a file, such
// as read, write or execute.
type OpenMode byte

// FileMode is the mode and permissions of a file.
type FileMode uint32

// QidType specifies the filetype in Qid structs, such as a regular file,
// directory or auth.
type QidType byte

// Qid is the servers unique file identification.
type Qid struct {
	Type QidType

	// Version describes the version of the file. It is usually incremented
	// every time the file is changed.
	Version uint32

	// Path is a unique identifier for the file within a file server.
	Path uint64
}

func (*Qid) EncodedLength() int {
	return 13
}

func (q *Qid) Decode(r io.Reader) error {
	var err error
	if q.Type, err = ReadQidType(r); err != nil {
		return err
	}
	if q.Version, err = ReadUint32(r); err != nil {
		return err
	}
	if q.Path, err = ReadUint64(r); err != nil {
		return err
	}
	return nil
}

func (q *Qid) Encode(w io.Writer) error {
	var err error
	if err = WriteQidType(w, q.Type); err != nil {
		return err
	}
	if err = WriteUint32(w, q.Version); err != nil {
		return err
	}
	if err = WriteUint64(w, q.Path); err != nil {
		return err
	}
	return nil
}

// Stat is a directory entry, providing detailed information of a file. It is
// called "Dir" in many other implementations.
type Stat struct {
	// Type is reserved for kernel use.
	Type uint16

	// Dev is reserved for kernel use.
	Dev uint32

	// Qid is the Qid for the file.
	Qid Qid

	// Mode is the permissions and mode of the file.
	Mode FileMode

	// Atime is the last access time of the file.
	Atime uint32

	// Mtime is the last modification time of the file.
	Mtime uint32

	// Length is the length of the file, commonly 0 for directories.
	Length uint64

	// Name is the name of the file.
	Name string

	// UID is the username of the owning user.
	UID string

	// GID is the group name of the owning group.
	GID string

	// MUID is the user who last modified the file.
	MUID string
}

func (s *Stat) EncodedLength() int {
	return 2 + 2 + 4 + 13 + 4 + 4 + 4 + 8 + 8 + len(s.Name) + len(s.UID) + len(s.GID) + len(s.MUID)
}

func (s *Stat) Decode(r io.Reader) error {
	var err error

	// We have no use of this length
	if _, err = ReadUint16(r); err != nil {
		return err
	}

	if s.Type, err = ReadUint16(r); err != nil {
		return err
	}
	if s.Dev, err = ReadUint32(r); err != nil {
		return err
	}
	if err = s.Qid.Decode(r); err != nil {
		return err
	}
	if s.Mode, err = ReadFileMode(r); err != nil {
		return err
	}
	if s.Atime, err = ReadUint32(r); err != nil {
		return err
	}
	if s.Mtime, err = ReadUint32(r); err != nil {
		return err
	}
	if s.Length, err = ReadUint64(r); err != nil {
		return err
	}
	if s.Name, err = ReadString(r); err != nil {
		return err
	}
	if s.UID, err = ReadString(r); err != nil {
		return err
	}
	if s.GID, err = ReadString(r); err != nil {
		return err
	}
	if s.MUID, err = ReadString(r); err != nil {
		return err
	}
	return nil
}

func (s *Stat) Encode(w io.Writer) error {
	var err error

	l := uint16(s.EncodedLength() - 2)

	if err = WriteUint16(w, l); err != nil {
		return err
	}
	if err = WriteUint16(w, s.Type); err != nil {
		return err
	}
	if err = WriteUint32(w, s.Dev); err != nil {
		return err
	}
	if err = s.Qid.Encode(w); err != nil {
		return err
	}
	if err = WriteFileMode(w, s.Mode); err != nil {
		return err
	}
	if err = WriteUint32(w, s.Atime); err != nil {
		return err
	}
	if err = WriteUint32(w, s.Mtime); err != nil {
		return err
	}
	if err = WriteUint64(w, s.Length); err != nil {
		return err
	}
	if err = WriteString(w, s.Name); err != nil {
		return err
	}
	if err = WriteString(w, s.UID); err != nil {
		return err
	}
	if err = WriteString(w, s.GID); err != nil {
		return err
	}
	if err = WriteString(w, s.MUID); err != nil {
		return err
	}

	return nil
}

//
// Message type structs and the encode/decode methods below.
//

// VersionRequest is used to inform the server of the maximum size it intends
// to send or it can receive, as well as the maximum protocol version
// supported. The tag used for this request must be NOTAG.
type VersionRequest struct {
	Tag Tag

	// MaxSize is the suggested absolute maximum message size for the
	// connection. The final negotiated value must be honoured.
	MaxSize uint32

	// Version is the suggested maximum protocol version for the connection.
	Version string
}

func (vr *VersionRequest) GetTag() Tag {
	return vr.Tag
}

func (vr *VersionRequest) SetTag(t Tag) {
	vr.Tag = t
}

func (vr *VersionRequest) EncodedLength() int {
	return 2 + 4 + 2 + len(vr.Version)
}

func (vr *VersionRequest) Decode(r io.Reader) error {
	var err error
	if vr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if vr.MaxSize, err = ReadUint32(r); err != nil {
		return err
	}
	if vr.Version, err = ReadString(r); err != nil {
		return err
	}
	return nil
}

func (vr *VersionRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, vr.Tag); err != nil {
		return err
	}
	if err = WriteUint32(w, vr.MaxSize); err != nil {
		return err
	}
	if err = WriteString(w, vr.Version); err != nil {
		return err
	}
	return nil
}

// VersionResponse is used to inform the client of maximum size and version,
// taking the clients VersionRequest into consideration. MaxSize in the reply
// must not be larger than MaxSize in the request, and the version must
// likewise be equal to or lower than the one in the requst.
type VersionResponse struct {
	Tag Tag

	// MaxSize is the negotiated maximum message size for the connection. This value must be honoured.
	MaxSize uint32

	// Version is the negotiated protocol version, or "unknown" if negotiation failed.
	Version string
}

func (vr *VersionResponse) GetTag() Tag {
	return vr.Tag
}

func (vr *VersionResponse) SetTag(t Tag) {
	vr.Tag = t
}

func (vr *VersionResponse) EncodedLength() int {
	return 2 + 4 + 2 + len(vr.Version)
}

func (vr *VersionResponse) Decode(r io.Reader) error {
	var err error
	if vr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if vr.MaxSize, err = ReadUint32(r); err != nil {
		return err
	}
	if vr.Version, err = ReadString(r); err != nil {
		return err
	}
	return nil
}

func (vr *VersionResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, vr.Tag); err != nil {
		return err
	}
	if err = WriteUint32(w, vr.MaxSize); err != nil {
		return err
	}
	if err = WriteString(w, vr.Version); err != nil {
		return err
	}
	return nil
}

// AuthRequest is used to request and authentication protocol connection from
// the server. The AuthFid can be used to read/write the authentication
// protocol. The protocol itself is not part of 9P2000.
type AuthRequest struct {
	Tag Tag

	// AuthFid is the fid to be used for authentication protocol.
	AuthFid Fid

	// Username is the user to authenticate as.
	Username string

	// Service is the service to authenticate access to.
	Service string
}

func (ar *AuthRequest) GetTag() Tag {
	return ar.Tag
}

func (ar *AuthRequest) SetTag(t Tag) {
	ar.Tag = t
}

func (ar *AuthRequest) EncodedLength() int {
	return 2 + 4 + 2 + len(ar.Username) + 2 + len(ar.Service)
}

func (ar *AuthRequest) Decode(r io.Reader) error {
	var err error
	if ar.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if ar.AuthFid, err = ReadFid(r); err != nil {
		return err
	}
	if ar.Username, err = ReadString(r); err != nil {
		return err
	}
	if ar.Service, err = ReadString(r); err != nil {
		return err
	}
	return nil
}

func (ar *AuthRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, ar.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, ar.AuthFid); err != nil {
		return err
	}
	if err = WriteString(w, ar.Username); err != nil {
		return err
	}
	if err = WriteString(w, ar.Service); err != nil {
		return err
	}
	return nil
}

// AuthResponse is used to acknowledge the authentication protocol connection,
// and to return the matching Qid.
type AuthResponse struct {
	Tag Tag

	// AuthQid is the Qid representing the special authentication file.
	AuthQid Qid
}

func (ar *AuthResponse) GetTag() Tag {
	return ar.Tag
}

func (ar *AuthResponse) SetTag(t Tag) {
	ar.Tag = t
}

func (*AuthResponse) EncodedLength() int {
	return 2 + 13
}

func (ar *AuthResponse) Decode(r io.Reader) error {
	var err error
	if ar.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if err = ar.AuthQid.Decode(r); err != nil {
		return err
	}
	return nil
}

func (ar *AuthResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, ar.Tag); err != nil {
		return err
	}
	if err = ar.AuthQid.Encode(w); err != nil {
		return err
	}
	return nil
}

// AttachRequest is used to establish a connection to a service as a user, and
// attach a fid to the root of the service.
type AttachRequest struct {
	Tag Tag

	// Fid is the fid that will be assigned the root node.
	Fid Fid

	// AuthFid is the fid of the previously executed authentication protocol, or
	// NOFID is the service does not need authentication.
	AuthFid Fid

	// Username is the user the connection will operate as.
	Username string

	// Service is the service that will be accessed.
	Service string
}

func (ar *AttachRequest) GetTag() Tag {
	return ar.Tag
}

func (ar *AttachRequest) SetTag(t Tag) {
	ar.Tag = t
}

func (ar *AttachRequest) EncodedLength() int {
	return 2 + 4 + 4 + 2 + len(ar.Username) + 2 + len(ar.Service)
}

func (ar *AttachRequest) Decode(r io.Reader) error {
	var err error
	if ar.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if ar.Fid, err = ReadFid(r); err != nil {
		return err
	}
	if ar.AuthFid, err = ReadFid(r); err != nil {
		return err
	}
	if ar.Username, err = ReadString(r); err != nil {
		return err
	}
	if ar.Service, err = ReadString(r); err != nil {
		return err
	}
	return nil
}

func (ar *AttachRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, ar.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, ar.Fid); err != nil {
		return err
	}
	if err = WriteFid(w, ar.AuthFid); err != nil {
		return err
	}
	if err = WriteString(w, ar.Username); err != nil {
		return err
	}
	if err = WriteString(w, ar.Service); err != nil {
		return err
	}
	return nil
}

// AttachResponse acknowledges an attach.
type AttachResponse struct {
	Tag Tag

	// Qid is the qid of the root node.
	Qid Qid
}

func (ar *AttachResponse) GetTag() Tag {
	return ar.Tag
}

func (ar *AttachResponse) SetTag(t Tag) {
	ar.Tag = t
}

func (*AttachResponse) EncodedLength() int {
	return 2 + 13
}

func (ar *AttachResponse) Decode(r io.Reader) error {
	var err error
	if ar.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if err = ar.Qid.Decode(r); err != nil {
		return err
	}
	return nil
}

func (ar *AttachResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, ar.Tag); err != nil {
		return err
	}
	if err = ar.Qid.Encode(w); err != nil {
		return err
	}
	return nil
}

// ErrorResponse is used when the server wants to report and error with the
// request. There is no ErrorRequest, as such a thing would not make sense.
type ErrorResponse struct {
	Tag Tag

	// Error is the error string.
	Error string
}

func (er *ErrorResponse) GetTag() Tag {
	return er.Tag
}

func (er *ErrorResponse) SetTag(t Tag) {
	er.Tag = t
}

func (er *ErrorResponse) EncodedLength() int {
	return 2 + 2 + len(er.Error)
}

func (er *ErrorResponse) Decode(r io.Reader) error {
	var err error
	if er.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if er.Error, err = ReadString(r); err != nil {
		return err
	}
	return nil
}

func (er *ErrorResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, er.Tag); err != nil {
		return err
	}
	if err = WriteString(w, er.Error); err != nil {
		return err
	}
	return nil
}

// FlushRequest is used to cancel a pending request. The flushed tag can be
// used after a response have been received.
type FlushRequest struct {
	Tag Tag

	// OldTag is the tag of the request to cancel.
	OldTag Tag
}

func (fr *FlushRequest) GetTag() Tag {
	return fr.Tag
}

func (fr *FlushRequest) SetTag(t Tag) {
	fr.Tag = t
}

func (*FlushRequest) EncodedLength() int {
	return 2 + 2
}

func (fr *FlushRequest) Decode(r io.Reader) error {
	var err error
	if fr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if fr.OldTag, err = ReadTag(r); err != nil {
		return err
	}
	return nil
}

func (fr *FlushRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, fr.Tag); err != nil {
		return err
	}
	if err = WriteTag(w, fr.OldTag); err != nil {
		return err
	}
	return nil
}

// FlushResponse is used to indicate a successful flush. Do note that
// FlushResponse have a peculiar behaviour when multiple flushes are pending.
type FlushResponse struct {
	Tag Tag
}

func (fr *FlushResponse) GetTag() Tag {
	return fr.Tag
}

func (fr *FlushResponse) SetTag(t Tag) {
	fr.Tag = t
}

func (*FlushResponse) EncodedLength() int {
	return 2
}

func (fr *FlushResponse) Decode(r io.Reader) error {
	var err error
	if fr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	return nil
}

func (fr *FlushResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, fr.Tag); err != nil {
		return err
	}
	return nil
}

// WalkRequest is used to walk into directories, starting from the current fid.
// All but the last name must be directories. If the walk succeeds, the file is
// assigned to NewFid.
type WalkRequest struct {
	Tag Tag

	// Fid is the fid to walk from.
	Fid Fid

	// NewFid is the fid to assign the successful walk to.
	NewFid Fid

	// Names are the names to try.
	Names []string
}

func (wr *WalkRequest) GetTag() Tag {
	return wr.Tag
}

func (wr *WalkRequest) SetTag(t Tag) {
	wr.Tag = t
}

func (wr *WalkRequest) EncodedLength() int {
	x := 0
	for i := range wr.Names {
		x += 2 + len(wr.Names[i])
	}
	return 2 + 4 + 4 + 2 + x
}

func (wr *WalkRequest) Decode(r io.Reader) error {
	var err error
	if wr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if wr.Fid, err = ReadFid(r); err != nil {
		return err
	}
	if wr.NewFid, err = ReadFid(r); err != nil {
		return err
	}
	var arr uint16
	if arr, err = ReadUint16(r); err != nil {
		return err
	}
	wr.Names = make([]string, arr)
	for i := 0; i < int(arr); i++ {
		if wr.Names[i], err = ReadString(r); err != nil {
			return err
		}
	}
	return nil
}

func (wr *WalkRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, wr.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, wr.Fid); err != nil {
		return err
	}
	if err = WriteFid(w, wr.NewFid); err != nil {
		return err
	}
	if err = WriteUint16(w, uint16(len(wr.Names))); err != nil {
		return err
	}
	for i := range wr.Names {
		if err = WriteString(w, wr.Names[i]); err != nil {
			return err
		}
	}
	return nil
}

// WalkResponse returns the qids for each successfully walked element. If the
// walk is successful, the amount of qids will be identical to the amount of
// names.
type WalkResponse struct {
	Tag Tag

	// Qids are the qids for the successfully walked files.
	Qids []Qid
}

func (wr *WalkResponse) GetTag() Tag {
	return wr.Tag
}

func (wr *WalkResponse) SetTag(t Tag) {
	wr.Tag = t
}

func (wr *WalkResponse) EncodedLength() int {
	return 2 + 2 + 13*len(wr.Qids)
}

func (wr *WalkResponse) Decode(r io.Reader) error {
	var err error
	if wr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	var arr uint16
	if arr, err = ReadUint16(r); err != nil {
		return err
	}
	wr.Qids = make([]Qid, arr)
	for i := 0; i < int(arr); i++ {
		if err = wr.Qids[i].Decode(r); err != nil {
			return err
		}
	}
	return nil
}

func (wr *WalkResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, wr.Tag); err != nil {
		return err
	}
	if err = WriteUint16(w, uint16(len(wr.Qids))); err != nil {
		return err
	}
	for i := range wr.Qids {
		if err = wr.Qids[i].Encode(w); err != nil {
			return err
		}
	}
	return nil
}

// OpenRequest is used to open a fid for reading/writing/executing.
type OpenRequest struct {
	Tag Tag

	// Fid is the file to open.
	Fid Fid

	// Mode is the mode to open file under.
	Mode OpenMode
}

func (or *OpenRequest) GetTag() Tag {
	return or.Tag
}

func (or *OpenRequest) SetTag(t Tag) {
	or.Tag = t
}

func (*OpenRequest) EncodedLength() int {
	return 2 + 4 + 1
}

func (or *OpenRequest) Decode(r io.Reader) error {
	var err error
	if or.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if or.Fid, err = ReadFid(r); err != nil {
		return err
	}
	if or.Mode, err = ReadOpenMode(r); err != nil {
		return err
	}
	return nil
}

func (or *OpenRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, or.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, or.Fid); err != nil {
		return err
	}
	if err = WriteOpenMode(w, or.Mode); err != nil {
		return err
	}
	return nil
}

// OpenResponse returns the qid of the file, as well as iounit, which is a
// read/write size that is guaranteed to be sucessfully written/read, or 0 for
// no such guarantee.
type OpenResponse struct {
	Tag Tag

	// Qid is the qid of the opened file.
	Qid Qid

	// IOUnit is the maximum amount of data that can be read/written by a single
	// call, or 0 for no specification.
	IOUnit uint32
}

func (or *OpenResponse) GetTag() Tag {
	return or.Tag
}

func (or *OpenResponse) SetTag(t Tag) {
	or.Tag = t
}

func (*OpenResponse) EncodedLength() int {
	return 2 + 13 + 4
}

func (or *OpenResponse) Decode(r io.Reader) error {
	var err error
	if or.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if err = or.Qid.Decode(r); err != nil {
		return err
	}
	if or.IOUnit, err = ReadUint32(r); err != nil {
		return err
	}
	return nil
}

func (or *OpenResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, or.Tag); err != nil {
		return err
	}
	if err = or.Qid.Encode(w); err != nil {
		return err
	}
	if err = WriteUint32(w, or.IOUnit); err != nil {
		return err
	}
	return nil
}

// CreateRequest tries to create a file in the current directory with the
// provided permissions, and then open it with behaviour identical to
// OpenRequest. A directory is created by creating a file with the DMDIR
// permission bit set.
type CreateRequest struct {
	Tag Tag

	// Fid is the fid of the directory where the file should be created, but
	// upon successful creation and opening, it changes to the opened file.
	Fid Fid

	// Name is the name of the file to create.
	Name string

	// Permissions are the permissions and mode of the file to create.
	Permissions FileMode

	// Mode is the mode the file should be opened under.
	Mode OpenMode
}

func (cr *CreateRequest) GetTag() Tag {
	return cr.Tag
}

func (cr *CreateRequest) SetTag(t Tag) {
	cr.Tag = t
}

func (cr *CreateRequest) EncodedLength() int {
	return 2 + 4 + 2 + len(cr.Name) + 4 + 1
}

func (cr *CreateRequest) Decode(r io.Reader) error {
	var err error
	if cr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if cr.Fid, err = ReadFid(r); err != nil {
		return err
	}
	if cr.Name, err = ReadString(r); err != nil {
		return err
	}
	if cr.Permissions, err = ReadFileMode(r); err != nil {
		return err
	}
	if cr.Mode, err = ReadOpenMode(r); err != nil {
		return err
	}
	return nil
}

func (cr *CreateRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, cr.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, cr.Fid); err != nil {
		return err
	}
	if err = WriteString(w, cr.Name); err != nil {
		return err
	}
	if err = WriteFileMode(w, cr.Permissions); err != nil {
		return err
	}
	if err = WriteOpenMode(w, cr.Mode); err != nil {
		return err
	}
	return nil
}

// CreateResponse returns the qid of the file, as well as iounit, which is a
// read/write size that is guaranteed to be sucessfully written/read, or 0 for
// no such guarantee.
type CreateResponse struct {
	Tag Tag

	// Qid is the qid of the opened file.
	Qid Qid

	// IOUnit is the maximum amount of data that can be read/written by a single
	// call, or 0 for no specification.
	IOUnit uint32
}

func (cr *CreateResponse) GetTag() Tag {
	return cr.Tag
}

func (cr *CreateResponse) SetTag(t Tag) {
	cr.Tag = t
}

func (*CreateResponse) EncodedLength() int {
	return 2 + 13 + 4
}

func (cr *CreateResponse) Decode(r io.Reader) error {
	var err error
	if cr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if err = cr.Qid.Decode(r); err != nil {
		return err
	}
	if cr.IOUnit, err = ReadUint32(r); err != nil {
		return err
	}
	return nil
}

func (cr *CreateResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, cr.Tag); err != nil {
		return err
	}
	if err = cr.Qid.Encode(w); err != nil {
		return err
	}
	if err = WriteUint32(w, cr.IOUnit); err != nil {
		return err
	}
	return nil
}

// ReadRequest is used to read data from an open file.
type ReadRequest struct {
	Tag Tag

	// Fid is the fid of the file to read.
	Fid Fid

	// Offset is used to continue a previous read or to seek in the file.
	Offset uint64

	// Count is the maximum amount of byte requested.
	Count uint32
}

func (rr *ReadRequest) GetTag() Tag {
	return rr.Tag
}

func (rr *ReadRequest) SetTag(t Tag) {
	rr.Tag = t
}

func (*ReadRequest) EncodedLength() int {
	return 2 + 4 + 8 + 4
}

func (rr *ReadRequest) Decode(r io.Reader) error {
	var err error
	if rr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if rr.Fid, err = ReadFid(r); err != nil {
		return err
	}
	if rr.Offset, err = ReadUint64(r); err != nil {
		return err
	}
	if rr.Count, err = ReadUint32(r); err != nil {
		return err
	}
	return nil
}

func (rr *ReadRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, rr.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, rr.Fid); err != nil {
		return err
	}
	if err = WriteUint64(w, rr.Offset); err != nil {
		return err
	}
	if err = WriteUint32(w, rr.Count); err != nil {
		return err
	}
	return nil
}

// ReadResponse returns read data.
type ReadResponse struct {
	Tag Tag

	// Data is the data that was read.
	Data []byte
}

func (rr *ReadResponse) GetTag() Tag {
	return rr.Tag
}

func (rr *ReadResponse) SetTag(t Tag) {
	rr.Tag = t
}

func (rr *ReadResponse) EncodedLength() int {
	return 2 + 4 + len(rr.Data)
}

func (rr *ReadResponse) Decode(r io.Reader) error {
	var err error
	if rr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	var l uint32
	if l, err = ReadUint32(r); err != nil {
		return err
	}
	rr.Data = make([]byte, l)
	if err = read(r, rr.Data); err != nil {
		return err
	}
	return nil
}

func (rr *ReadResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, rr.Tag); err != nil {
		return err
	}
	if err = WriteUint32(w, uint32(len(rr.Data))); err != nil {
		return err
	}
	if err = write(w, rr.Data); err != nil {
		return err
	}
	return nil
}

// WriteRequest is used to write to an open file.
type WriteRequest struct {
	Tag Tag

	// Fid is the file to write to.
	Fid Fid

	// Offset is used to continue a previous write or to seek.
	Offset uint64

	// Data is the data to write.
	Data []byte
}

func (wr *WriteRequest) GetTag() Tag {
	return wr.Tag
}

func (wr *WriteRequest) SetTag(t Tag) {
	wr.Tag = t
}

func (wr *WriteRequest) EncodedLength() int {
	return 2 + 4 + 8 + 4 + len(wr.Data)
}

func (wr *WriteRequest) Decode(r io.Reader) error {
	var err error
	if wr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if wr.Fid, err = ReadFid(r); err != nil {
		return err
	}
	if wr.Offset, err = ReadUint64(r); err != nil {
		return err
	}
	var count uint32
	if count, err = ReadUint32(r); err != nil {
		return err
	}
	wr.Data = make([]byte, count)
	if err = read(r, wr.Data); err != nil {
		return err
	}
	return nil
}

func (wr *WriteRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, wr.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, wr.Fid); err != nil {
		return err
	}
	if err = WriteUint64(w, wr.Offset); err != nil {
		return err
	}
	if err = WriteUint32(w, uint32(len(wr.Data))); err != nil {
		return err
	}
	if err = write(w, wr.Data); err != nil {
		return err
	}
	return nil
}

// WriteResponse informs of how much data was written.
type WriteResponse struct {
	Tag Tag

	// Count is the amount of written data.
	Count uint32
}

func (wr *WriteResponse) GetTag() Tag {
	return wr.Tag
}

func (wr *WriteResponse) SetTag(t Tag) {
	wr.Tag = t
}

func (*WriteResponse) EncodedLength() int {
	return 2 + 4
}

func (wr *WriteResponse) Decode(r io.Reader) error {
	var err error
	if wr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if wr.Count, err = ReadUint32(r); err != nil {
		return err
	}
	return nil
}

func (wr *WriteResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, wr.Tag); err != nil {
		return err
	}
	if err = WriteUint32(w, wr.Count); err != nil {
		return err
	}
	return nil
}

// ClunkRequest is used to clear a fid, allowing it to be reused.
type ClunkRequest struct {
	Tag Tag

	// Fid is the fid to clunk.
	Fid Fid
}

func (cr *ClunkRequest) GetTag() Tag {
	return cr.Tag
}

func (cr *ClunkRequest) SetTag(t Tag) {
	cr.Tag = t
}

func (*ClunkRequest) EncodedLength() int {
	return 2 + 4
}

func (cr *ClunkRequest) Decode(r io.Reader) error {
	var err error
	if cr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if cr.Fid, err = ReadFid(r); err != nil {
		return err
	}
	return nil
}

func (cr *ClunkRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, cr.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, cr.Fid); err != nil {
		return err
	}
	return nil
}

// ClunkResponse indicates a successful clunk.
type ClunkResponse struct {
	Tag Tag
}

func (cr *ClunkResponse) GetTag() Tag {
	return cr.Tag
}

func (cr *ClunkResponse) SetTag(t Tag) {
	cr.Tag = t
}

func (*ClunkResponse) EncodedLength() int {
	return 2
}

func (cr *ClunkResponse) Decode(r io.Reader) error {
	var err error
	if cr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	return nil
}

func (cr *ClunkResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, cr.Tag); err != nil {
		return err
	}
	return nil
}

// RemoveRequest is used to clunk a fid and remove the file if possible.
type RemoveRequest struct {
	Tag Tag

	// Fid is the fid to clunk and potentially remove.
	Fid Fid
}

func (rr *RemoveRequest) GetTag() Tag {
	return rr.Tag
}

func (rr *RemoveRequest) SetTag(t Tag) {
	rr.Tag = t
}

func (*RemoveRequest) EncodedLength() int {
	return 2 + 4
}

func (rr *RemoveRequest) Decode(r io.Reader) error {
	var err error
	if rr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if rr.Fid, err = ReadFid(r); err != nil {
		return err
	}
	return nil
}

func (rr *RemoveRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, rr.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, rr.Fid); err != nil {
		return err
	}
	return nil
}

// RemoveResponse indicates a successful clunk, but not necessarily a successful remove.
type RemoveResponse struct {
	Tag Tag
}

func (rr *RemoveResponse) GetTag() Tag {
	return rr.Tag
}

func (rr *RemoveResponse) SetTag(t Tag) {
	rr.Tag = t
}

func (*RemoveResponse) EncodedLength() int {
	return 2
}

func (rr *RemoveResponse) Decode(r io.Reader) error {
	var err error
	if rr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	return nil
}

func (rr *RemoveResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, rr.Tag); err != nil {
		return err
	}
	return nil
}

// StatRequest is used to retrieve the Stat struct of a file
type StatRequest struct {
	Tag Tag

	// Fid is the fid to retrieve Stat for.
	Fid Fid
}

func (sr *StatRequest) GetTag() Tag {
	return sr.Tag
}

func (sr *StatRequest) SetTag(t Tag) {
	sr.Tag = t
}

func (*StatRequest) EncodedLength() int {
	return 2 + 4
}

func (sr *StatRequest) Decode(r io.Reader) error {
	var err error
	if sr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if sr.Fid, err = ReadFid(r); err != nil {
		return err
	}
	return nil
}

func (sr *StatRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, sr.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, sr.Fid); err != nil {
		return err
	}
	return nil
}

// StatResponse contains the Stat struct of a file.
type StatResponse struct {
	Tag Tag

	// Stat is the requested Stat struct.
	Stat Stat
}

func (sr *StatResponse) GetTag() Tag {
	return sr.Tag
}

func (sr *StatResponse) SetTag(t Tag) {
	sr.Tag = t
}

func (sr *StatResponse) EncodedLength() int {
	return 2 + 2 + sr.Stat.EncodedLength()
}

func (sr *StatResponse) Decode(r io.Reader) error {
	var err error
	if sr.Tag, err = ReadTag(r); err != nil {
		return err
	}

	// We don't need this
	if _, err = ReadUint16(r); err != nil {
		return err
	}

	if err = sr.Stat.Decode(r); err != nil {
		return err
	}
	return nil
}

func (sr *StatResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, sr.Tag); err != nil {
		return err
	}

	if err = WriteUint16(w, uint16(sr.Stat.EncodedLength())); err != nil {
		return err
	}

	if err = sr.Stat.Encode(w); err != nil {
		return err
	}

	return nil
}

// WriteStatRequest attempts to apply a Stat struct to a file. This requires a
// combination of write permissions to the file as well as to the parent
// directory, depending on the properties changed. Properties can be set to "no
// change" values, which for strings are empty strings, and for integral values
// are the maximum unsigned value of their respective types. The write is
// either completely successful with all changes applied, or failed with no
// changes applied. The server must not perform a partial application of the
// Stat structure.
type WriteStatRequest struct {
	Tag Tag

	// Fid is the file to modify the Stat struct for.
	Fid Fid

	// Stat is the Stat struct to apply
	Stat Stat
}

func (wsr *WriteStatRequest) GetTag() Tag {
	return wsr.Tag
}

func (wsr *WriteStatRequest) SetTag(t Tag) {
	wsr.Tag = t
}

func (wsr *WriteStatRequest) EncodedLength() int {
	return 2 + 4 + 2 + wsr.Stat.EncodedLength()
}

func (wsr *WriteStatRequest) Decode(r io.Reader) error {
	var err error
	if wsr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	if wsr.Fid, err = ReadFid(r); err != nil {
		return err
	}

	// We don't need the stat size
	if _, err = ReadUint16(r); err != nil {
		return err
	}

	if err = wsr.Stat.Decode(r); err != nil {
		return err
	}
	return nil
}

func (wsr *WriteStatRequest) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, wsr.Tag); err != nil {
		return err
	}
	if err = WriteFid(w, wsr.Fid); err != nil {
		return err
	}
	if err = WriteUint16(w, uint16(wsr.Stat.EncodedLength())); err != nil {
		return err
	}
	if err = wsr.Stat.Encode(w); err != nil {
		return err
	}
	return nil
}

// WriteStatResponse indicates a successful application of a Stat structure.
type WriteStatResponse struct {
	Tag Tag
}

func (wsr *WriteStatResponse) GetTag() Tag {
	return wsr.Tag
}

func (wsr *WriteStatResponse) SetTag(t Tag) {
	wsr.Tag = t
}

func (*WriteStatResponse) EncodedLength() int {
	return 2
}

func (wsr *WriteStatResponse) Decode(r io.Reader) error {
	var err error
	if wsr.Tag, err = ReadTag(r); err != nil {
		return err
	}
	return nil
}

func (wsr *WriteStatResponse) Encode(w io.Writer) error {
	var err error
	if err = WriteTag(w, wsr.Tag); err != nil {
		return err
	}
	return nil
}
