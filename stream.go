package job

import "io"

const (
	stdOutErrName = "stdouterr"
)

type Stream interface {
	Source() StreamInput
	Dest() StreamOutput
}

type StreamInput interface {
	Name() string
	Reader() io.Reader
	SetReader(io.Reader)
}

type StreamOutput interface {
	Name() string
	Writer() io.Writer
	SetWriter(io.Writer)
}

type ioStream struct {
	source *ioInput
	dest   *ioOutput
}

func (s *ioStream) Source() StreamInput { return s.source }
func (s *ioStream) Dest() StreamOutput  { return s.dest }

type ioInput struct {
	name   string
	reader io.Reader
}

func (i *ioInput) Name() string          { return i.name }
func (i *ioInput) Reader() io.Reader     { return i.reader }
func (i *ioInput) SetReader(r io.Reader) { i.reader = r }

type ioOutput struct {
	name   string
	writer io.Writer
}

func (o *ioOutput) Name() string          { return o.name }
func (o *ioOutput) Writer() io.Writer     { return o.writer }
func (o *ioOutput) SetWriter(w io.Writer) { o.writer = w }

func NewStdOutErrStream() Stream {
	return &ioStream{
		source: &ioInput{name: stdOutErrName},
		dest:   &ioOutput{name: stdOutErrName},
	}
}

func NewNamedStream(name string) Stream {
	return &ioStream{
		source: &ioInput{name: name},
		dest:   &ioOutput{name: name},
	}
}
