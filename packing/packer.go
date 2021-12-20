package packing

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/kwstars/ktcp/message"
	"github.com/spf13/cast"
)

const (
	UnknownType = iota
	OKType
	ErrType
)

// Packer is a generic interface to pack and unpack message packet.
type Packer interface {
	// Pack packs Message into the packet to be written.
	Pack(msg *message.Message) ([]byte, error)

	// Unpack unpacks the message packet from reader,
	// returns the Message, and error if error occurred.
	Unpack(reader io.Reader) (*message.Message, error)
}

var _ Packer = &DefaultPacker{}

// NewDefaultPacker create a *DefaultPacker with initial field value.
func NewDefaultPacker() *DefaultPacker {
	return &DefaultPacker{
		MaxDataSize: 1 << 10 << 10, // 1MB
	}
}

// DefaultPacker is the default Packer used in session.
// Treats the packet with the format:
//
// dataSize(4)|id(4)|data(n)
//
// | segment    | type   | size    | remark                  |
// | ---------- | ------ | ------- | ----------------------- |
// | `dataSize` | uint32 | 4       | the size of `data` only |
// | `id`       | uint32 | 4       |                         |
// | `flag`     | uint16 | 2       |                         |
// | `data`     | []byte | dynamic |                         |
// .
type DefaultPacker struct {
	// MaxDataSize represents the max size of `data`
	MaxDataSize int
}

func (d *DefaultPacker) bytesOrder() binary.ByteOrder {
	return binary.LittleEndian
}

// Pack implements the Packer Pack method.
func (d *DefaultPacker) Pack(msg *message.Message) ([]byte, error) {
	dataSize := len(msg.Data)
	buffer := make([]byte, 4+4+2+dataSize)
	d.bytesOrder().PutUint32(buffer[:4], uint32(dataSize)) // write dataSize
	id, err := cast.ToUint32E(msg.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid type of msg.ID: %s", err)
	}
	d.bytesOrder().PutUint32(buffer[4:8], id)        // write id
	d.bytesOrder().PutUint16(buffer[8:10], msg.Flag) // write flag
	copy(buffer[10:], msg.Data)                      // write data
	return buffer, nil
}

// Unpack implements the Packer Unpack method.
// Unpack returns the msg whose ID is type of int.
// So we need use int id to register routes.
func (d *DefaultPacker) Unpack(reader io.Reader) (*message.Message, error) {
	headerBuffer := make([]byte, 4+4+2)
	if _, err := io.ReadFull(reader, headerBuffer); err != nil {
		return nil, fmt.Errorf("read size and id err: %s", err)
	}
	dataSize := d.bytesOrder().Uint32(headerBuffer[:4])
	if d.MaxDataSize > 0 && int(dataSize) > d.MaxDataSize {
		return nil, fmt.Errorf("the dataSize %d is beyond the max: %d", dataSize, d.MaxDataSize)
	}
	id := d.bytesOrder().Uint32(headerBuffer[4:8])
	flag := d.bytesOrder().Uint16(headerBuffer[8:10])
	data := make([]byte, dataSize)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("read data err: %s", err)
	}
	msg := &message.Message{
		ID:   id,
		Flag: flag,
		Data: data,
	}
	return msg, nil
}
