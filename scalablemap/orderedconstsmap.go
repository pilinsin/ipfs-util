package scalablemap

import(
	"github.com/pilinsin/util"
	ipfs "github.com/pilinsin/ipfs-util"
)

type indexedValue struct{
	value []byte
	idx int
}
func newIndexedValue(val []byte, idx int) *indexedValue{
	return &indexedValue{val, idx}
}
func (iv *indexedValue) Marshal() []byte{
	miv := &struct{
		Val []byte
		Idx int
	}{iv.value, iv.idx}
	m, _ := util.Marshal(miv)
	return m
}
func (iv *indexedValue) Unmarshal(m []byte) error{
	miv := &struct{
		Val []byte
		Idx int
	}{}
	if err := util.Unmarshal(m, miv); err != nil{return err}

	iv.value = miv.Val
	iv.idx = miv.Idx
	return nil
}

type orderedScalableMap struct{
	bm baseMap
	cids   []string
	capacity int
	idx int
}
func newOrderedScalableMap(capacity int) IScalableMap{
	return &orderedScalableMap{
		bm: make(baseMap, capacity),
		cids: make([]string, 0),
		capacity: capacity,
		idx: 0,
	}
}
func (sm orderedScalableMap) Len(is *ipfs.IPFS) int {
	length := len(sm.bm)
	for _, cid := range sm.cids {
		bm := baseMap{}
		if err := bm.fromCid(cid, is); err != nil{
			return -1
		}
		length += len(bm)
	}
	return length
}
func (sm orderedScalableMap) Type() string{return "ordered-append-only-map"}
func (sm *orderedScalableMap) Append(key interface{}, value []byte, is *ipfs.IPFS) error{
	if _, ok := sm.ContainKey(key, is); ok {
		return util.NewError("append error: already contain key")
	}

	hash := keyToTypeHash(key, sm.Type())
	value = newIndexedValue(value, sm.idx).Marshal()
	sm.idx++
	sm.bm[hash] = value
	if len(sm.bm) >= sm.capacity {
		cid := sm.bm.toCid(is)
		sm.cids = append(sm.cids, cid)
		sm.bm = make(baseMap, sm.capacity)
		sm.idx = 0
	}
	return nil
}
func (sm orderedScalableMap) Next(is *ipfs.IPFS) <-chan []byte {
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		for _, cid := range sm.cids{
			bm := baseMap{}
			if err := bm.fromCid(cid, is); err != nil{return}
			values := make([][]byte, len(bm))
			for _, v := range bm{
				idxValue := &indexedValue{}
				if err := idxValue.Unmarshal(v); err != nil{return}
				values[idxValue.idx] = idxValue.value
			}
			for _, v := range values{
				ch <- v
			}
		}

		values := make([][]byte, len(sm.bm))
		for _, v := range sm.bm{
			idxValue := &indexedValue{}
			if err := idxValue.Unmarshal(v); err != nil{return}
			values[idxValue.idx] = idxValue.value
		}
		for _, v := range values{
			ch <- v
		}
	}()
	return ch
}
func (sm orderedScalableMap) NextKeyValue(is *ipfs.IPFS) <-chan *keyValue {
	ch := make(chan *keyValue)
	go func() {
		defer close(ch)
		for _, cid := range sm.cids{
			bm := baseMap{}
			if err := bm.fromCid(cid, is); err != nil{return}
			keys := make([]string, len(bm))
			values := make([][]byte, len(bm))
			for key, v := range bm{
				idxValue := &indexedValue{}
				if err := idxValue.Unmarshal(v); err != nil{return}
				keys[idxValue.idx] = key
				values[idxValue.idx] = idxValue.value
			}
			for idx, _ := range values{
				ch <- &keyValue{keys[idx], values[idx]}
			}
		}

		keys := make([]string, len(sm.bm))
		values := make([][]byte, len(sm.bm))
		for key, v := range sm.bm{
			idxValue := &indexedValue{}
			if err := idxValue.Unmarshal(v); err != nil{return}
			keys[idxValue.idx] = key
			values[idxValue.idx] = idxValue.value
		}
		for idx, _ := range values{
			ch <- &keyValue{keys[idx], values[idx]}
		}
	}()
	return ch
}
func (sm orderedScalableMap) containKey(key interface{}, tp string, is *ipfs.IPFS) ([]byte, bool){
	hash := keyToTypeHash(key, sm.Type())
	
	if v, ok := sm.bm[hash]; ok{return v, true}
	for _, cid := range sm.cids{
		bm := baseMap{}
		if err := bm.fromCid(cid, is); err != nil{return nil, false}
		if v, ok := bm[hash]; ok{return v, true}
	}
	return nil, false
}
func (sm orderedScalableMap) ContainKey(key interface{}, is *ipfs.IPFS) ([]byte, bool) {
	v, ok := sm.containKey(key, sm.Type(), is)
	if !ok{return nil, false}
	idxValue := &indexedValue{}
	if err := idxValue.Unmarshal(v); err != nil{return nil, false}

	return idxValue.value, true
}
func (sm orderedScalableMap) ContainCid(cid string, is *ipfs.IPFS) bool {
	m, err := ipfs.File.Get(cid, is)
	if err != nil{return false}
	sm0 := &orderedScalableMap{}
	if err := sm0.Unmarshal(m); err != nil{return false}

	if sm0.capacity != sm.capacity{return false}
	for _, cid0 := range sm0.cids{
		if ok := util.StrSliceContain(sm.cids, cid0); !ok{return false}
	}
	bMap0 := sm0.bm.toMap()
	if ok := util.MapContainMap(sm.bm.toMap(), bMap0); ok{return true}
	for _, cid := range sm.cids{
		bm := &baseMap{}
		if err := bm.fromCid(cid, is); err != nil{return false}
		if ok := util.MapContainMap(bm.toMap(), bMap0); ok{return true}
	}
	return false
}
func (sm *orderedScalableMap) Marshal() []byte{
	msm := &struct{
		Bm baseMap
		Cids   []string
		Cap int
		Idx int
	}{sm.bm, sm.cids, sm.capacity, sm.idx}
	m, _ := util.Marshal(msm)
	return m
}
func (sm *orderedScalableMap) Unmarshal(m []byte) error{
	msm := &struct{
		Bm baseMap
		Cids   []string
		Cap int
		Idx int
	}{}
	if err := util.Unmarshal(m, msm); err != nil{return err}

	sm.bm = msm.Bm
	sm.cids = msm.Cids
	sm.capacity = msm.Cap
	sm.idx = msm.Idx
	return nil
}