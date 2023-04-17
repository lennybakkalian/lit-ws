package litws

type PacketId uint16

const (
	SubscriptionCreate PacketId = 0x0001
	SubscriptionDelete          = 0x0002

	C2SSyncMapSubscribe   = 0x0101 // subscribe to a sync map
	C2SSyncMapUnsubscribe = 0x0102 // unsubscribes
	C2SSyncMapFetch       = 0x0103 // fetches part of the sync map
	S2CSyncMapData        = 0x0104 // sync map data (insert,update,delete)
)

type Packet struct {
	Id PacketId `json:"id"`
	// TrackingId is used for callbacks or subscriptions.
	TrackingId string      `json:"trackingId"`
	Payload    interface{} `json:"payload"`
}
type packetHandler func(p *Packet, c *Client)

type PC2SSyncMapSubscribe struct {
	Key string `json:"key"`
}

type PC2SSyncMapUnsubscribe struct {
	Key string `json:"key"`
}

type PC2SSyncMapFetch struct {
	Key string `json:"key"`

	StartAt   interface{} `json:"startAt"` // fetches from this value. (inclusive)
	OrderBy   string      `json:"orderBy"`
	OrderDesc bool        `json:"orderDesc"`
	Count     int         `json:"count"`
}

type PS2CSyncMapData struct {
	// New can contain id:nil, id:new or id:updated values.
	New    map[uint64]interface{} `json:"new"`
	Delete []uint64               `json:"delete"`
	Errors []string               `json:"errors"`
}
