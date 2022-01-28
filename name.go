package ipfs

import(
	"context"
	"time"
	"strings"

	cid "github.com/ipfs/go-cid"
	kstore "github.com/ipfs/go-ipfs-keystore"
	iface "github.com/ipfs/interface-go-ipfs-core"
	ipath "github.com/ipfs/interface-go-ipfs-core/path"
	options "github.com/ipfs/interface-go-ipfs-core/options"
	nsopts "github.com/ipfs/interface-go-ipfs-core/options/namesys"

	"github.com/pilinsin/util"
)

type name struct{
	api iface.CoreAPI
	ctx     context.Context
	kStore  kstore.Keystore
}
func (self *IPFS) Name() *name{
	return &name{self.api, self.ctx, self.kStore}
}
func (self *name) hasKey(kw string) bool {
	keys, _ := self.api.Key().List(self.ctx)
	for _, key := range keys {
		if key.Name() == kw {
			return true
		}
	}
	return false
}
func parseDuration(vt string) time.Duration {
	t, err := time.ParseDuration(vt)
	if err != nil {
		t, _ = time.ParseDuration("8760h")
	}
	return t
}
func (self *name) PublishWithKeyFile(pth ipath.Path, vt string, kFile *KeyFile) iface.IpnsEntry {
	t := parseDuration(vt)

	var kw string
	for {
		kw = util.GenUniqueID(50, 50)
		if ng := self.hasKey(kw); !ng {
			break
		}
	}
	self.kStore.Put(kw, kFile.keyFile)
	ipnsEntry, _ := self.api.Name().Publish(self.ctx, pth, options.Name.ValidTime(t), options.Name.Key(kw))
	self.kStore.Delete(kw)
	return ipnsEntry
}
func (self *name) Publish(pth ipath.Path, vt string, kw string) iface.IpnsEntry {
	t := parseDuration(vt)

	if kw == ""{kw = "self"}
	if kw != "self" && !self.hasKey(kw) {
		self.api.Key().Generate(self.ctx, kw, options.Key.Type("ed25519"))
	}
	ipnsEntry, _ := self.api.Name().Publish(self.ctx, pth, options.Name.ValidTime(t), options.Name.Key(kw))
	return ipnsEntry
}
func (self *name) Resolve(name string) (ipath.Path, error) {
	ctx, cancel := util.CancelTimerContext(10*time.Second)
	defer cancel()
	return self.api.Name().Resolve(ctx, name, options.Name.ResolveOption(nsopts.DhtRecordCount(1)))
}


type nameUtil struct{}
var Name nameUtil
func (self nameUtil) Publish(data []byte, kw string, is *IPFS) string {
	pth := is.File().Add(data, true)
	return is.Name().Publish(pth, "", kw).Name()
}
func (self nameUtil) PublishWithKeyFile(data []byte, kf *KeyFile, is *IPFS) string {
	pth := is.File().Add(data, true)
	return is.Name().PublishWithKeyFile(pth, "", kf).Name()
}
func (self nameUtil) PublishCid(cidStr string, kw string, is *IPFS) string {
	cid, err := cid.Decode(cidStr)
	if err != nil {
		return ""
	}
	pth := ipath.IpfsPath(cid)
	return is.Name().Publish(pth, "", kw).Name()
}
func (self nameUtil) PublishCidWithKeyFile(cidStr string, kf *KeyFile, is *IPFS) string {
	cid, err := cid.Decode(cidStr)
	if err != nil {
		return ""
	}
	pth := ipath.IpfsPath(cid)
	return is.Name().PublishWithKeyFile(pth, "", kf).Name()
}
func (self nameUtil) Get(ipnsName string, is *IPFS) ([]byte, error) {
	pth, err := is.Name().Resolve(ipnsName)
	if err != nil {
		return nil, err
	} else {
		return is.File().Get(pth)
	}
}
func (self nameUtil) GetCid(ipnsName string, is *IPFS) (string, error) {
	pth, err := is.Name().Resolve(ipnsName)
	if err != nil {
		return "", err
	} else {
		return strings.TrimPrefix(pth.String(), "/ipfs/"), nil
	}
}