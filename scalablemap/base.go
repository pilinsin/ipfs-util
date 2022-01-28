package scalablemap

import (
	"fmt"
	"github.com/pilinsin/util"
	"github.com/pilinsin/util/crypto"
	ipfs "github.com/pilinsin/ipfs-util"
)

func keyToTypeHash(key interface{}, tp string) string{
	kb := util.AnyStrToBytes64(fmt.Sprintln(key))
	tb := util.AnyStrToBytes64(tp)
	return util.AnyBytes64ToStr(crypto.Hash(kb, tb))
}

type keyValue struct {
	key   string
	value []byte
}
func (kv keyValue) Key() string {
	return kv.key
}
func (kv keyValue) Value() []byte {
	return kv.value
}

type baseMap map[string][]byte
func (bm baseMap) toMap() map[string][]byte{
	return map[string][]byte(bm)
}
func (bm baseMap) toCid(is *ipfs.IPFS) string{
	m, _ := util.Marshal(bm)
	return ipfs.File.Add(m, is)
}
func (bm *baseMap) fromCid(cid string, is *ipfs.IPFS) error{
	m, err := ipfs.File.Get(cid, is)
	if err != nil{return err}
	return util.Unmarshal(m, bm)
}

