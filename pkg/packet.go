package litws

type PacketId uint16

const (
	SubscriptionCreate PacketId = 0x0001
	SubscriptionDelete          = 0x0002

	C2SSyncMapSubscribe   = 0x0101 // subscribe to a sync map
	C2SSyncMapUnsubscribe = 0x0102 // unsubscribes
	C2SSyncMapFetch       = 0x0103 // fetches part of the sync map
)

type Packet[T any] struct {
	Id         PacketId
	TrackingId string // used for callbacks or subscriptions
	Payload    T
}
type packetHandler func(p *Packet[any], c *Client)

type PC2SSyncMapSubscribe struct {
	Key string `json:"key"`
}

type PC2SSyncMapUnsubscribe struct {
	Key string `json:"key"`
}

type PC2SSyncMapFetch struct {
	Key string `json:"key"`
}
