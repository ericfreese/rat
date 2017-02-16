package rat

import (
	"io"
	"sync"
)

type Stream interface {
	io.WriteCloser
	Bytes() []byte
	NewReader() StreamReader
}

type StreamReader interface {
	io.Reader
}

type streamReader struct {
	stream *stream
	off    int
	block  bool
}

func (r *streamReader) Read(p []byte) (n int, err error) {
	r.stream.cond.L.Lock()
	defer r.stream.cond.L.Unlock()

	if !r.stream.closed {
		for r.off+len(p) > len(r.stream.bytes) {
			r.stream.cond.Wait()

			if r.stream.closed {
				err = io.EOF
				break
			}
		}
	} else if r.off+len(p) >= len(r.stream.bytes) {
		err = io.EOF
	}

	n = copy(p, r.stream.bytes[r.off:])
	r.off = r.off + n
	return
}

type stream struct {
	bytes  []byte
	closed bool
	cond   *sync.Cond
}

func NewStream() Stream {
	s := &stream{}

	s.bytes = make([]byte, 0, 128)
	s.cond = &sync.Cond{L: &sync.Mutex{}}

	return s
}

func (s *stream) Write(p []byte) (n int, err error) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	s.bytes = append(s.bytes, p...)
	s.cond.Broadcast()

	return len(p), nil
}

func (s *stream) Close() error {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	s.closed = true
	s.cond.Broadcast()

	return nil
}

func (s *stream) Bytes() []byte {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	return s.bytes
}

func (s *stream) NewReader() StreamReader {
	r := &streamReader{}
	r.stream = s
	return r
}
