package scalablemap

import (
	ipfs "github.com/pilinsin/ipfs-util"
	"github.com/pilinsin/util"
)

type constScalableMap struct {
	bm       baseMap
	cids     map[string]struct{}
	capacity int
}

func newConstScalableMap(capacity int) IScalableMap {
	return &constScalableMap{
		bm:       make(baseMap, capacity),
		cids:     make(map[string]struct{}, 0),
		capacity: capacity,
	}
}
func (sm constScalableMap) Len(is *ipfs.IPFS) int {
	length := len(sm.bm)
	for cid, _ := range sm.cids {
		bm := baseMap{}
		if err := bm.fromCid(cid, is); err != nil {
			return -1
		}
		length += len(bm)
	}
	return length
}
func (sm constScalableMap) Type() string { return "append-only-map" }
func (sm *constScalableMap) Append(key interface{}, value []byte, is *ipfs.IPFS) error {
	if _, ok := sm.ContainKey(key, is); ok {
		return util.NewError("append error: already contain key")
	}

	hash := keyToTypeHash(key, sm.Type())
	sm.bm[hash] = value
	if len(sm.bm) >= sm.capacity {
		cid := sm.bm.toCid(is)
		sm.cids[cid] = struct{}{}
		sm.bm = make(baseMap, sm.capacity)
	}
	return nil
}
func (sm constScalableMap) Next(is *ipfs.IPFS) <-chan []byte {
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
func (sm constScalableMap) NextKeyValue(is *ipfs.IPFS) <-chan *keyValue {
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
func (sm constScalableMap) ContainKey(key interface{}, is *ipfs.IPFS) ([]byte, bool) {
	hash := keyToTypeHash(key, sm.Type())

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
func (sm constScalableMap) ContainCid(cid string, is *ipfs.IPFS) bool {
	m, err := ipfs.File.Get(cid, is)
	if err != nil {
		return false
	}
	sm0 := &constScalableMap{}
	if err := sm0.Unmarshal(m); err != nil {
		return false
	}

	if sm0.capacity != sm.capacity {
		return false
	}
	for cid0, _ := range sm0.cids {
		if _, ok := sm.cids[cid0]; !ok {
			return false
		}
	}
	bMap0 := sm0.bm.toMap()
	if ok := util.MapContainMap(sm.bm.toMap(), bMap0); ok {
		return true
	}
	for cid, _ := range sm.cids {
		bm := &baseMap{}
		if err := bm.fromCid(cid, is); err != nil {
			return false
		}
		if ok := util.MapContainMap(bm.toMap(), bMap0); ok {
			return true
		}
	}
	return false
}
func (sm *constScalableMap) Marshal() []byte {
	mScalableMap := &struct {
		Bm   baseMap
		Cids map[string]struct{}
		Cap  int
	}{sm.bm, sm.cids, sm.capacity}
	m, _ := util.Marshal(mScalableMap)
	return m
}
func (sm *constScalableMap) Unmarshal(m []byte) error {
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
