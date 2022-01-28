package ipfs

import (
	"context"
	"time"

	cid "github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	iface "github.com/ipfs/interface-go-ipfs-core"
	options "github.com/ipfs/interface-go-ipfs-core/options"
	ipath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/pilinsin/util"
)

type file struct {
	api iface.CoreAPI
	ctx context.Context
}

func (self *IPFS) File() *file {
	return &file{self.api, self.ctx}
}
func (self *file) Add(data []byte, pn bool) ipath.Resolved {
	file := files.NewBytesFile(data)
	pth, _ := self.api.Unixfs().Add(self.ctx, file, options.Unixfs.Pin(pn))
	return pth
}
func (self *file) Hash(data []byte) ipath.Resolved {
	file := files.NewBytesFile(data)
	pth, _ := self.api.Unixfs().Add(self.ctx, file, options.Unixfs.HashOnly(true))
	return pth
}

func (self *file) Get(pth ipath.Path) ([]byte, error) {
	ctx, cancel := util.CancelTimerContext(10 * time.Second)
	defer cancel()
	f, err := self.api.Unixfs().Get(ctx, pth)
	if err != nil {
		return nil, err
	} else {
		return ipfsFileNodeToBytes(f)
	}
}

type fileUtil struct{}

var File fileUtil

func (self fileUtil) Hash(data []byte, is *IPFS) string {
	return is.File().Hash(data).Cid().String()
}
func (self fileUtil) Add(data []byte, is *IPFS) string {
	return is.File().Add(data, true).Cid().String()
}
func (self fileUtil) Get(cidStr string, is *IPFS) ([]byte, error) {
	cid, err := cid.Decode(cidStr)
	if err != nil {
		return nil, err
	}
	pth := ipath.IpfsPath(cid)
	return is.File().Get(pth)
}
