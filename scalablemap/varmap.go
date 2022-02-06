package scalablemap

import (
	ipfs "github.com/pilinsin/ipfs-util"
	"github.com/pilinsin/util"
)

type varScalableMap struct {
	bm       baseMap
	cids     map[string]struct{}
	capacity int
}

func newVarScalableMap(capacity int) IScalableMap {
	return &varScalableMap{
		bm:       make(baseMap, capacity),
		cids:     make(map[string]struct{}, 0),
		capacity: capacity,
	}
}
func (sm varScalableMap) Len() int {
	return len(sm.bm) + sm.capacity * len(sm.cids)
}
func (sm varScalableMap) Type() string { return "variable-map" }
func (sm *varScalableMap) Append(key interface{}, value []byte, is *ipfs.IPFS) error {
	hash := keyToTypeHash(key, sm.Type())

	for cid, _ := range sm.cids{
		bm := baseMap{}
		if err := bm.fromCid(cid, is); err != nil {
			return err
		}
		if _, ok := bm[hash]; ok{
			bm[hash] = value
			cid2 := bm.toCid(is)
			delete(sm.cids, cid)
			sm.cids[cid2] = struct{}{}
			return nil
		}
	}

	_, update := sm.bm[hash]
	sm.bm[hash] = value
	if !update && len(sm.bm) >= sm.capacity {
		cid := sm.bm.toCid(is)
		sm.cids[cid] = struct{}{}
		sm.bm = make(baseMap, sm.capacity)
	}
	return nil
}
func (sm varScalableMap) Next(is *ipfs.IPFS) <-chan []byte {
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		for cid, _ := range sm.cids {
			bm := baseMap{}
			if err := bm.fromCid(cid, is); err != nil {
				return
			}
			for _, v := range bm {
				ch <- v
			}
		}
		for _, v := range sm.bm {
			ch <- v
		}
	}()
	return ch
}
func (sm varScalableMap) NextKeyValue(is *ipfs.IPFS) <-chan *keyValue {
	ch := make(chan *keyValue)
	go func() {
		defer close(ch)
		for cid, _ := range sm.cids {
			bm := baseMap{}
			if err := bm.fromCid(cid, is); err != nil {
				return
			}
			for k, v := range bm {
				ch <- &keyValue{k, v}
			}
		}
		for k, v := range sm.bm {
			ch <- &keyValue{k, v}
		}
	}()
	return ch
}
func (sm varScalableMap) ContainKey(key interface{}, is *ipfs.IPFS) ([]byte, bool) {
	hash := keyToTypeHash(key, sm.Type())
	return sm.ContainKeyHash(hash, is)
}
func (sm varScalableMap) ContainKeyHash(hash string, is *ipfs.IPFS) ([]byte, bool){
	if v, ok := sm.bm[hash]; ok {
		return v, true
	}
	for cid, _ := range sm.cids {
		bm := baseMap{}
		if err := bm.fromCid(cid, is); err != nil {
			return nil, false
		}
		if v, ok := bm[hash]; ok {
			return v, true
		}
	}
	return nil, false
}
func (sm varScalableMap) ContainMap(contained IScalableMap, is *ipfs.IPFS) bool {
	sm0, ok := contained.(*varScalableMap)
	if !ok{return false}

	if sm0.capacity != sm.capacity {
		return false
	}

	for kv := range sm0.NextKeyValue(is){
		hash := kv.Key()
		v0 := kv.Value()
		v, ok := sm.ContainKeyHash(hash, is)
		if !ok || !util.ConstTimeBytesEqual(v, v0){
			return false
		}
	}
	return true
}
func (sm *varScalableMap) Marshal() []byte {
	mScalableMap := &struct {
		Bm   baseMap
		Cids map[string]struct{}
		Cap  int
	}{sm.bm, sm.cids, sm.capacity}
	m, _ := util.Marshal(mScalableMap)
	return m
}
func (sm *varScalableMap) Unmarshal(m []byte) error {
	mScalableMap := &struct {
		Bm   baseMap
		Cids map[string]struct{}
		Cap  int
	}{}
	if err := util.Unmarshal(m, mScalableMap); err != nil {
		return err
	}

	sm.bm = mScalableMap.Bm
	sm.cids = mScalableMap.Cids
	sm.capacity = mScalableMap.Cap
	return nil
}
