package scalablemap

import (
	ipfs "github.com/pilinsin/ipfs-util"
)

type IScalableMap interface{
	Len(is *ipfs.IPFS) int
	Type() string
	Next(is *ipfs.IPFS) <-chan []byte
	NextKeyValue(is *ipfs.IPFS) <-chan *keyValue
	Append(key interface{}, value []byte, is *ipfs.IPFS) error
	ContainKey(key interface{}, is *ipfs.IPFS) ([]byte, bool)
	ContainCid(cid string, is *ipfs.IPFS) bool
	Marshal() []byte
	Unmarshal(m []byte) error
}
func NewScalableMap(mode string, capacity int) IScalableMap{
	switch mode {
	case "ordered":
		return newOrderedScalableMap(capacity)
	default:
		return newConstScalableMap(capacity)
	}
}
func UnmarshalScalableMap(mode string, m []byte) (IScalableMap, error){
	switch mode {
	case "ordered":
		sm := &orderedScalableMap{}
		err := sm.Unmarshal(m)
		return sm, err
	default:
		sm := &constScalableMap{}
		err := sm.Unmarshal(m)
		return sm, err
	}
}
