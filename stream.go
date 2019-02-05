package job

import "io"

const (
	stdOutErrName = "stdouterr"
)

type Stream interface {
	Name() string
	Source() io.Reader
	SetSource(io.Reader)
	Dest() io.Writer
	SetDest(io.Writer)
}

type ioStream struct {
	name   string
	source io.Reader
	dest   io.Writer
}

func (s *ioStream) Name() string {
	return s.name
}

func (s *ioStream) Source() io.Reader {
	return s.source
}

func (s *ioStream) SetSource(r io.Reader) {
	s.source = r
}

func (s *ioStream) Dest() io.Writer {
	return s.dest
}

func (s *ioStream) SetDest(w io.Writer) {
	s.dest = w
}

func NewStdOutErrStream() Stream {
	return NewNamedStream(stdOutErrName)
}

func NewNamedStream(name string) Stream {
	return &ioStream{name: name}
}
