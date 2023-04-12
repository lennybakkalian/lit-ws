package litws

import "time"

type PacketId uint16

const (
	SubscriptionCreate PacketId = iota
	SubscriptionDelete
)

type Packet interface {
	Name() string
	Id() PacketId
}
type packetHandler func(msg Decoder, c *Client)
type Decoder interface {
	Decode(val interface{}) error
	Time() time.Time
}
