package litws

import "log"

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
		C2SSyncMapSubscribe: s.handleSyncMapSubscribe,
	})
}

func (s *Synchronizer) RegisterSyncMap(key string, syncMap *SyncMap[uint64, any]) {
	if s.syncMap[key] != nil {
		panic("syncMap already registered: " + key)
	}
	s.syncMap[key] = syncMap
}

func (s *Synchronizer) handleSyncMapSubscribe(p *Packet[any], c *Client) {
	payload := p.Payload.(*PC2SSyncMapSubscribe)
	syncMap, ok := s.syncMap[payload.Key]
	if !ok {
		log.Printf("syncMap not found: %s", payload.Key)
		return
	}
	s.syncMapSubs = append(s.syncMapSubs, newSynchronizedSubscription(c, syncMap))
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

func (ss *SynchronizedSubscription) onSet(key uint64, value any) {

}

func (ss *SynchronizedSubscription) onDelete(key uint64, _ any) {

}

func (ss *SynchronizedSubscription) onUpdate(key uint64, value any) {

}

func (ss *SynchronizedSubscription) onFetch(packet *Packet[PC2SSyncMapFetch]) {

}
