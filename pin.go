package ipfs

import(
	iface "github.com/ipfs/interface-go-ipfs-core"
	ipath "github.com/ipfs/interface-go-ipfs-core/path"
)

type pin struct{
	api iface.CoreAPI
	ctx     context.Context
}
func (self *IPFS) Pin() *pin{
	return &pin{self.api, self.ctx}
}
func (self *pin) Add(pth ipath.Path) error{
	return self.api.Pin().Add(self.ctx, pth)
}
func (self *pin) Ls() (<-chan iface.Pin, error){
	return self.api.Pin().Ls(self.ctx)
}
func (self *pin) IsPinned(pth ipath.Path) (string, bool, error){
	return self.api.Pin().IsPinned(self.ctx, pth)
}
func (self *pin) Rm(pth ipath.Path) error{
	return self.api.Pin().Rm(self.ctx, pth)
}
func (self *pin) Update(from, to ipath.Path) error{
	return self.api.Pin().Update(self.ctx, from, to)
}
