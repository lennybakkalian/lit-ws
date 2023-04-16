package litws

import (
	"log"
)

type SynchronizedSubscription struct {
	client  *Client
	syncMap *SyncMap[uint64, any]
}

type Synchronizer struct {
	lws *Litws

	syncMap     map[string]*SyncMap[uint64, any]
	syncMapSubs []*SynchronizedSubscription
}

func (lws *Litws) NewSynchronizer() *Synchronizer {
	return &Synchronizer{
		lws:         lws,
		syncMap:     map[string]*SyncMap[uint64, any]{},
		syncMapSubs: []*SynchronizedSubscription{},
	}
}

func (s *Synchronizer) init() {
	s.lws.RegisterPacketHandlers(map[uint64]packetHandler{
		C2SSyncMapSubscribe:   s.handleSyncMapSubscribe,
		C2SSyncMapUnsubscribe: s.handleSyncMapUnsubscribe,
		C2SSyncMapFetch:       s.handleSyncMapFetch,
	})
}

func (s *Synchronizer) RegisterSyncMap(key string, syncMap *SyncMap[uint64, any]) {
	if s.syncMap[key] != nil {
		panic("syncMap already registered: " + key)
	}
	if syncMap.SortFunc == nil {
		panic("syncMap sortFunc must not be nil")
	}
	s.syncMap[key] = syncMap
}

func (s *Synchronizer) handleSyncMapSubscribe(p *Packet, c *Client) {
	payload := p.Payload.(*PC2SSyncMapSubscribe)
	syncMap, ok := s.syncMap[payload.Key]
	if !ok {
		log.Printf("syncMap not found: %s", payload.Key)
		return
	}
	s.syncMapSubs = append(s.syncMapSubs, newSynchronizedSubscription(c, syncMap))
}

func (s *Synchronizer) handleSyncMapUnsubscribe(p *Packet, c *Client) {
	payload := p.Payload.(*PC2SSyncMapUnsubscribe)
	sm := s.syncMap[payload.Key]
	for i, sub := range s.syncMapSubs {
		if sub.client == c && sub.syncMap == sm {
			sub.unsubscribe()
			s.syncMapSubs = append(s.syncMapSubs[:i], s.syncMapSubs[i+1:]...)
			break
		}
	}
}

func (s *Synchronizer) handleSyncMapFetch(p *Packet, c *Client) {
	payload := p.Payload.(*PC2SSyncMapFetch)
	sm := s.syncMap[payload.Key]
	for _, sub := range s.syncMapSubs {
		if sub.client == c && sub.syncMap == sm {
			sub.onFetch(p)
			break
		}
	}
}

// SynchronizedSubscription
func newSynchronizedSubscription(c *Client, syncMap *SyncMap[uint64, any]) *SynchronizedSubscription {
	ss := &SynchronizedSubscription{c, syncMap}
	syncMap.addEventListener("set", ss.onSet)
	syncMap.addEventListener("delete", ss.onDelete)
	syncMap.addEventListener("update", ss.onUpdate)
	return ss
}
func (ss *SynchronizedSubscription) unsubscribe() {
	ss.syncMap.removeEventListener("set", ss.onSet)
	ss.syncMap.removeEventListener("delete", ss.onDelete)
	ss.syncMap.removeEventListener("update", ss.onUpdate)
}

func (ss *SynchronizedSubscription) onSet(values map[uint64]*any) {

}

func (ss *SynchronizedSubscription) onDelete(values map[uint64]*any) {

}

func (ss *SynchronizedSubscription) onUpdate(values map[uint64]*any) {

}

func (ss *SynchronizedSubscription) onFetch(packet *Packet) {
	payload := packet.Payload.(*PC2SSyncMapFetch)
	respond := PS2CSyncMapData{New: map[uint64]interface{}{}}

	sorted := ss.syncMap.GetSortedList(payload.OrderBy, payload.OrderDesc)
	fillList := false
	for _, k := range sorted {
		if !fillList && less(payload.StartAt, ss.syncMap.ValueByField(payload.OrderBy, k.Value)) {
			fillList = true
		}
		if fillList {
			respond.New[k.Id] = k.Value
		}
		if len(respond.New) >= payload.Count {
			break
		}
	}

	ss.client.send <- &Packet{Id: S2CSyncMapData, Payload: respond}
}

func less(a, b interface{}) bool {
	switch a.(type) {
	case uint64:
		return a.(uint64) < b.(uint64)
	case string:
		return a.(string) < b.(string)
	case float64:
		return a.(float64) < b.(float64)
	default:
		panic("unknown type for less")
	}
	return false
}
