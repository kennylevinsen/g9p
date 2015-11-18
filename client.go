package g9p

import (
	"errors"
	"io"
	"sync"

	"github.com/joushou/g9p/protocol"
)

// Errors
var (
	ErrTagInUse        = errors.New("tag already in use")
	ErrNoSuchTag       = errors.New("tag does not exist")
	ErrInvalidResponse = errors.New("invalid response")
)

// Client implements a 9P2000 client on a ReadWriter.
type Client struct {
	rw        io.ReadWriter
	queueLock sync.RWMutex
	queue     map[protocol.Tag]chan protocol.Message
	writeLock sync.Mutex
	nextTag   protocol.Tag
}

// NextTag retrieves the next valid tag.
func (c *Client) NextTag() protocol.Tag {
	t := c.nextTag
	c.nextTag++
	if c.nextTag == protocol.NOTAG {
		c.nextTag++
	}
	return t
}

func (c *Client) getTag(t protocol.Tag) (chan protocol.Message, error) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()
	if _, ok := c.queue[t]; ok {
		return nil, ErrTagInUse
	}

	ch := make(chan protocol.Message, 1)
	c.queue[t] = ch
	return ch, nil
}

func (c *Client) handleResponse(d protocol.Message) error {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()
	t := d.GetTag()
	if ch, ok := c.queue[t]; ok {
		ch <- d
		delete(c.queue, t)
		return nil
	}
	return ErrNoSuchTag
}

func (c *Client) write(t protocol.Tag, d protocol.Message) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	if err := protocol.Encode(c.rw, d); err != nil {
		if _, ok := c.queue[t]; ok {
			delete(c.queue, t)

		}
		return err
	}
	return nil
}

func (c *Client) send(d protocol.Message) (protocol.Message, error) {
	t := d.GetTag()
	ch, err := c.getTag(t)
	if err != nil {
		return nil, err
	}
	c.write(t, d)
	resp := <-ch
	if resp == nil {
		return nil, ErrFlushed
	}

	if e, ok := resp.(*protocol.ErrorResponse); ok {
		return nil, errors.New(e.Error)
	}
	return resp, nil
}

func (c *Client) flush(t protocol.Tag) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()
	if ch, ok := c.queue[t]; ok {
		ch <- nil
		delete(c.queue, t)
	}
}
func (c *Client) Version(r *protocol.VersionRequest) (*protocol.VersionResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.VersionResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Auth(r *protocol.AuthRequest) (*protocol.AuthResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.AuthResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Attach(r *protocol.AttachRequest) (*protocol.AttachResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.AttachResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Flush(r *protocol.FlushRequest) (*protocol.FlushResponse, error) {
	// TODO(kl): Handle of multiple flushes on a single request.
	t := r.OldTag
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.FlushResponse); ok {
		c.flush(t)
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Walk(r *protocol.WalkRequest) (*protocol.WalkResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.WalkResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Open(r *protocol.OpenRequest) (*protocol.OpenResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.OpenResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Create(r *protocol.CreateRequest) (*protocol.CreateResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.CreateResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Read(r *protocol.ReadRequest) (*protocol.ReadResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.ReadResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Write(r *protocol.WriteRequest) (*protocol.WriteResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.WriteResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Clunk(r *protocol.ClunkRequest) (*protocol.ClunkResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.ClunkResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Remove(r *protocol.RemoveRequest) (*protocol.RemoveResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.RemoveResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) Stat(r *protocol.StatRequest) (*protocol.StatResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.StatResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

func (c *Client) WriteStat(r *protocol.WriteStatRequest) (*protocol.WriteStatResponse, error) {
	resp, err := c.send(r)
	if err != nil {
		return nil, err
	}

	if resp, ok := resp.(*protocol.WriteStatResponse); ok {
		return resp, nil
	}
	return nil, ErrInvalidResponse
}

// Start starts serving the responses for the client.
func (c *Client) Start() error {
	defer func() {
		if closer, ok := c.rw.(io.Closer); ok {
			closer.Close()
		}
	}()

	for {
		var (
			size uint32
			mt   protocol.MessageType
			err  error
		)

		if size, mt, err = protocol.DecodeHdr(c.rw); err != nil {
			return err
		}

		limiter := &io.LimitedReader{R: c.rw, N: int64(size) - protocol.HeaderSize}

		var r protocol.Message
		if r, err = protocol.MessageTypeToMessage(mt); err != nil {
			return err
		}
		if err = r.Decode(limiter); err != nil {
			return err
		}

		c.handleResponse(r)

	}
}

// Stop stops a client.
func (c *Client) Stop() {
	// TODO(kl): Add more robust stop.
	defer func() {
		if closer, ok := c.rw.(io.Closer); ok {
			closer.Close()
		}
	}()
}

// NewClient returns a new client serving the provided ReadWriter.
func NewClient(rw io.ReadWriter) *Client {
	return &Client{
		rw:    rw,
		queue: make(map[protocol.Tag]chan protocol.Message),
	}
}
