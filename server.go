package g9p

import (
	"io"
	"net"
	"sync"

	"github.com/joushou/g9p/protocol"
)

type Server struct {
	Handler   Handler
	RW        io.ReadWriter
	writeLock sync.Mutex
}

func (s *Server) handleResponse(tag protocol.Tag, d protocol.Message, e error) {
	if e == ErrFlushed {
		return
	}

	if e != nil {
		d = &protocol.ErrorResponse{Tag: tag, Error: e.Error()}
	}

	s.writeLock.Lock()
	defer s.writeLock.Unlock()

	protocol.Encode(s.RW, d)
}

func (s *Server) Start() error {
	for {
		var (
			size uint32
			mt   protocol.MessageType
			err  error
		)

		if size, mt, err = protocol.DecodeHdr(s.RW); err != nil {
			return err
		}

		// This LimitedReader is not a necessity, but simply a sanity check.
		limiter := &io.LimitedReader{R: s.RW, N: int64(size) - protocol.HeaderSize}

		switch mt {
		case protocol.Tversion:
			r := &protocol.VersionRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}

			go func(r *protocol.VersionRequest) {
				tag := r.Tag
				res, err := s.Handler.Version(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Tauth:
			r := &protocol.AuthRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}

			go func(r *protocol.AuthRequest) {
				tag := r.Tag
				res, err := s.Handler.Auth(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Tattach:
			r := &protocol.AttachRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.AttachRequest) {
				tag := r.Tag
				res, err := s.Handler.Attach(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Tflush:
			r := &protocol.FlushRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.FlushRequest) {
				tag := r.Tag
				res, err := s.Handler.Flush(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Twalk:
			r := &protocol.WalkRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.WalkRequest) {
				tag := r.Tag
				res, err := s.Handler.Walk(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Topen:
			r := &protocol.OpenRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.OpenRequest) {
				tag := r.Tag
				res, err := s.Handler.Open(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Tcreate:
			r := &protocol.CreateRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.CreateRequest) {
				tag := r.Tag
				res, err := s.Handler.Create(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Tread:
			r := &protocol.ReadRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.ReadRequest) {
				tag := r.Tag
				res, err := s.Handler.Read(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Twrite:
			r := &protocol.WriteRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.WriteRequest) {
				tag := r.Tag
				res, err := s.Handler.Write(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Tclunk:
			r := &protocol.ClunkRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.ClunkRequest) {
				tag := r.Tag
				res, err := s.Handler.Clunk(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Tremove:
			r := &protocol.RemoveRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.RemoveRequest) {
				tag := r.Tag
				res, err := s.Handler.Remove(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Tstat:
			r := &protocol.StatRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.StatRequest) {
				tag := r.Tag
				res, err := s.Handler.Stat(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		case protocol.Twstat:
			r := &protocol.WriteStatRequest{}
			if err = r.Decode(limiter); err != nil {
				return err
			}
			go func(r *protocol.WriteStatRequest) {
				tag := r.Tag
				res, err := s.Handler.WriteStat(r)
				if res != nil {
					res.Tag = tag
				}
				s.handleResponse(tag, res, err)
			}(r)
		default:
			return protocol.ErrUnknownMessageType
		}
	}
}

func Serve(rw io.ReadWriter, handler Handler) error {
	s := Server{
		Handler: handler,
		RW:      rw,
	}

	err := s.Start()
	if c, ok := s.RW.(io.Closer); ok {
		c.Close()
	}
	return err
}

func ServeListener(l net.Listener, handler func() Handler) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go Serve(conn, handler())
	}
}
