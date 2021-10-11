package utils

type PipeArrayBuffer struct {
	view   []byte
	buffer []byte
}

func NewPipeArrayBuffer(bytes []byte, len uint64) *PipeArrayBuffer {
	return &PipeArrayBuffer{
		view:   make([]byte, len),
		buffer: bytes,
	}
}

func (p *PipeArrayBuffer) ByteLength() int {
	return len(p.view)
}

func (p *PipeArrayBuffer) Read(num int64) []byte {
	bytes := p.view[0:num]
	p.view = p.view[num:]
	return bytes
}

func (p *PipeArrayBuffer) readUint8() uint8 {
	bytes := p.view[0]
	p.view = p.view[1:]
	return uint8(bytes)
}

func (p *PipeArrayBuffer) write(bytes []byte) {
	p.view = append(p.buffer, bytes...)
}

func (p *PipeArrayBuffer) end() bool {
	return len(p.view) == 0
}
