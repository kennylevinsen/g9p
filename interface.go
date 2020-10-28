package g9p

import (
	"errors"

	"github.com/kennylevinsen/g9p/protocol"
)

var (
	// ErrFlushed indicates that a request was flushed.
	ErrFlushed = errors.New("request flushed")
)

// Handler is the interface exposed by g9p's 9P2000 protocol handling
// mechanisms, both for server and client applications. A client implements
// this interface, allowing a user to call the methods on the Handler, whereas
// a server accepts and operates on a device implementing the Handler
// interface. The result is that a call to a Handler method on the client
// results in the very same call to the Handler on the server, returned values
// communicated back to give the effect of a completely transparent
// communication layer. Due to the dual purpose of this interface, one could
// also use g9p's server and client to implement a proxy, as a client can be
// plugged directly into a server as Handler.
//
// The server should take care of setting the tag of the response to that of
// the request, but the user of a client is responsible for setting the initial
// request tag. This could easily be taken care of by the client
// implementation, but doing so would make it impossible for a user to know the
// currently used tag, and thereby allow for the request to be cancelled with a
// Flush.
//
// The docs for this interface can either be read as description of behaviour
// from the side of a client, or the contract that a server must honour when
// implementing the interface. The docs below try to explain the important
// aspects of "high-level" behaviour, rather than the intricate details of the
// protocol. For more info, see http://man.cat-v.org/plan_9/5/intro as well as
// the pages in http://man.cat-v.org/plan_9/5/
type Handler interface {

	// Version is part of the initial protocol negotiation. It must be the first
	// action on a connection, and the tag must be set to protocol.NOTAG.  The
	// request contains the suggested maximum message size, including all
	// encoding but without transport layer protocols such as TCP, as well as
	// the suggested protocol, in this case 9P2000. The returned response
	// contains the negotiated maximum message size, which must be smaller than
	// or equal to the suggested size, and the protocol, which must be older or
	// equal to the suggested protocol. If the server is unable or unwilling to
	// negotiate a protocol based on the client's demands, it will return a
	// protocol name "unknown".
	Version(*protocol.VersionRequest) (*protocol.VersionResponse, error)

	// Auth is used to execute an authentication protocol not directly specified
	// by 9P2000. The request contains a fid (AuthFid), as well as username and
	// requested service name. The fid will represent an authentication file,
	// where the authentication protocol can be read and written to using the
	// usual 9P2000 protocol. The response contains an Qid for this
	// authentication file of type QTAUTH. After completed authentication, the
	// AuthFid can be used in one or more subsequent Attach messages to be
	// granted access to the services. An error must be returned if the service
	// does not require authentication, in which case the client can proceed
	// directly to Attach.
	Auth(*protocol.AuthRequest) (*protocol.AuthResponse, error)

	// Attach is used to set up the request fid to be connected to the root of
	// the requested service. It takes the fid to be prepared, as well as an
	// afid from a previous authentication request if an authentication protocol
	// has been executed (otherwise set to protocol.NOFID), the username the
	// user wants to attach as, as well as the name of the requested service.
	// The response contains the Qid representing the root of the service. If
	// the user does not have permission to attach or the fid is already taken,
	// an error is returned.
	Attach(*protocol.AttachRequest) (*protocol.AttachResponse, error)

	// Flush is used to drop existing requests. When Flush returns, the tag is
	// free to be used again. A response may arrive before Flush returns, in
	// which case it must be handled appropriately, as it might reflect a change
	// of state on the server. The request contains the tag of the request that
	// should be flushed. Flush may be called multiple times for a given tag, in
	// which case the server is only required to respond to the last flush.
	// Flush is mainly used to interrupt long-running reads or writes.
	Flush(*protocol.FlushRequest) (*protocol.FlushResponse, error)

	// Walk is used to attempt entering directories from the provided list, one
	// element at a time, starting from the element of the provided fid. The
	// request takes the fid to operate on, a newfid to assign the final element
	// to, and a list of names to walk. The fid must not have been opened by
	// Open or Create. The list of names must not exceed 16 names. To walk more
	// than 16 elements, walk more than once. The response contains a list of
	// qids for each successfully walked element. A walk with an empty list, or
	// a walk to ".", results in newfid being attached to the same file as the
	// initial fid. If the walk list is not empty, then walk returns an error if
	// the initial fid did not represent a directory. If a name from the walk
	// list is not found, or a name that is not the last in the list is not a
	// directory, the walk will terminate prematurely without modifying fid or
	// newfid, returning the already generated list of qids for the existing
	// elements. If the walk is successful, the list of qids will be as long as
	// the list of names, and newfid will represent the last element.
	Walk(*protocol.WalkRequest) (*protocol.WalkResponse, error)

	// Open is used to open a fid for manipulation. The request takes the fid
	// and the mode for opening. The response contains the qid and an optional
	// iounit, which is a measure of largest message that can be read or written
	// successfully without being prematurely terminated, or 0 for no
	// guarantees.
	Open(*protocol.OpenRequest) (*protocol.OpenResponse, error)

	// Create is used to create and open an element. The request takes a the fid
	// of the directory in which you want to create an element, the name you
	// want to create, the permissions you want to create it with and the mode
	// you want to open it with. Create will fail with an error if the current
	// fid is not a directory, or you do not have write permissions to the
	// directory. A file is created by default, but a directory can be created
	// by setting DMDIR as permission. On successful create, the fid is changed
	// to point to the created file. The semantics after file creation is
	// identical to Open, with the response containing the qid and iounit.
	Create(*protocol.CreateRequest) (*protocol.CreateResponse, error)

	// Read is used to read from an element. The request contains the fid to
	// read from, the offset to start reading from and the maximum wanted
	// bytecount. The fid must be opened for reading. If iounit was specified on
	// open, reads less than or equal to iounit bytes are guaranteed to be read
	// by the request. A larger count than iounit, or no iounit guarantee means
	// that a read will return any amount of bytes up to the requested count.
	// Read on directories return a list of protocol.Stat, encoded end-to-end,
	// one for each element in the directory.  A directory may only be read with
	// offset set to 0, or the previous offset + the previous count. That is,
	// seeking to anything but offset 0 is illegal. The response contains the
	// successfully read bytes, fewer than or equal to count.
	Read(*protocol.ReadRequest) (*protocol.ReadResponse, error)

	// Write is used to write to an element. The request contains the fid to
	// write from, the offset and data to write. Writing to a directory is
	// illegal. The fid must be opened for writing. If iounit was specified on
	// open, writes less than or equal to iounit bytes is guaranteed to be
	// written by the request. A larger write than iounit, or no iounit
	// guarantee means that a write will write any amount of bytes up to the
	// provided amount. The response contains the written bytes.
	Write(*protocol.WriteRequest) (*protocol.WriteResponse, error)

	// Clunk is used to invalidate a fid after use. The request takes the fid to
	// invalidate. If the fid was opened with ORCLOSE, the element represented
	// by the fid is also attempted removed. Once a fid has been clunked, it can
	// be reused, even if clunk returns an error. The response is empty, but
	// indicates a succssful clunk.
	Clunk(*protocol.ClunkRequest) (*protocol.ClunkResponse, error)

	// Remove is used to clunk a fid, and remove the file represented by the
	// fid. The request takes the fid of the element to clunk and remove. It is
	// equivalent to clunking a fid opened with ORCLOSE. The fid will be
	// clunked, even if remove fails. Removing a fid requires writing permission
	// to the parent directory. It is correct to consider remove to be a clunk
	// with the side effect of removing the file if permissions allow. The
	// response is empty.
	Remove(*protocol.RemoveRequest) (*protocol.RemoveResponse, error)

	// Stat is used to return the protocol.Stat structure for the element
	// represented by the fid. The request takes the fid of the element to stat.
	// The response contains the protocol.Stat structure for the element.
	Stat(*protocol.StatRequest) (*protocol.StatResponse, error)

	// WriteStat is used to modify the protocol.Stat structure of the element
	// represented by the fid. The request takes the fid of the element to
	// modify, and the protocol.Stat structure to apply. Either all the
	// requested changes are applied, or none are applied. A partial apply will
	// not occur. Special "don't touch" values can be set: empty strings and
	// maximum unsigned value of integrals. A special-case WriteStat will occur
	// if all values are set to "don't touch", which is meant to guarantee that
	// the file is committed to storage before a response is sent. It should
	// logically be interpretted as "make the state of the file exactly what it
	// claims to be". The response is empty.
	WriteStat(*protocol.WriteStatRequest) (*protocol.WriteStatResponse, error)
}
