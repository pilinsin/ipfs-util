package ipfs

import (
	"context"
	"path/filepath"
	"strings"

	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	unixfs "github.com/ipfs/go-unixfs"
	iface "github.com/ipfs/interface-go-ipfs-core"
	options "github.com/ipfs/interface-go-ipfs-core/options"
	ipath "github.com/ipfs/interface-go-ipfs-core/path"

	"github.com/pilinsin/util"
)

type link struct {
	Name string
	Size uint64
	Path ipath.Resolved
}

type object struct {
	api iface.CoreAPI
	ctx context.Context
}

func (self *IPFS) Object() *object {
	return &object{self.api, self.ctx}
}
func (self *object) NewRoot() ipath.Resolved {
	nd, _ := self.api.Object().New(self.ctx, options.Object.Type("unixfs-dir"))
	pbnd := nd.(*dag.ProtoNode)
	return ipath.IpfsPath(pbnd.Cid())
}
func (self *object) Get(pth ipath.Path) (ipld.Node, error) {
	return self.api.Object().Get(self.ctx, pth)
}
func (self *object) Links(pth ipath.Path) []link {
	ls, _ := self.api.Object().Links(self.ctx, pth)
	links := make([]link, len(ls))
	for idx, lnk := range ls {
		links[idx] = link{lnk.Name, lnk.Size, ipath.IpfsPath(lnk.Cid)}
	}
	return links
}
func (self *object) Stat(pth ipath.Path) (*iface.ObjectStat, error) {
	return self.api.Object().Stat(self.ctx, pth)
}
func (self *object) AddLink(root, child ipath.Path, path string) ipath.Resolved {
	pth, _ := self.api.Object().AddLink(self.ctx, root, path, child, options.Object.Create(true))
	return pth
}
func (self *object) RmLink(root ipath.Path, path string) ipath.Resolved {
	pth, _ := self.api.Object().RmLink(self.ctx, root, path)
	return pth
}

type objectUtil struct{}

var Object objectUtil

type fileSystem struct {
	obj           *object
	root, nowPath ipath.Resolved
	nowPathStr    string
}

func (self objectUtil) NewFileSystem(is *IPFS) *fileSystem {
	rt := is.Object().NewRoot()
	now := ipath.IpfsPath(rt.Cid())
	return &fileSystem{is.Object(), rt, now, "/"}
}
func (self *fileSystem) Add(pth ipath.Path, to string) {
	to = toAbsPath(self.nowPathStr, to)
	self.root = self.obj.AddLink(self.root, pth, to)
	self.nowPathStr = toAbsPath(self.nowPathStr, self.nowPathStr)
	self.nowPath, _ = self.findPath(self.nowPathStr)
}
func (self *fileSystem) Get(pathName string) (ipath.Resolved, error) {
	pathName = toAbsPath(self.nowPathStr, pathName)
	return self.findPath(pathName)
}
func (self *fileSystem) GetAllFiles(root ipath.Path) []link {
	files := make([]link, 0)
	if ok := isLinked(root, self.obj); !ok {
		return files
	}
	for _, link := range self.obj.Links(root) {
		if ok := isDir(link.Path, self.obj); !ok {
			files2 := self.GetAllFiles(link.Path)
			files = append(files, files2...)
		} else {
			files = append(files, link)
		}
	}
	return files
}
func (self *fileSystem) Cd(to string) {
	to = toAbsPath(self.nowPathStr, to)
	pth, err := self.findPath(to)
	if err == nil && isDir(pth, self.obj) {
		self.nowPath = pth
		self.nowPathStr = to
	}
}
func (self *fileSystem) Cp(from, to string) {
	from = toAbsPath(self.nowPathStr, from)
	to = toAbsPath(self.nowPathStr, to)
	if from == to {
		return
	}
	pth, err := self.findPath(from)
	if err != nil {
		return
	}

	self.root = self.obj.AddLink(self.root, pth, to)
	self.nowPathStr = toAbsPath(self.nowPathStr, self.nowPathStr)
	self.nowPath, _ = self.findPath(self.nowPathStr)
}
func (self *fileSystem) Mv(from, to string) {
	from = toAbsPath(self.nowPathStr, from)
	to = toAbsPath(self.nowPathStr, to)
	if from == "" || from == to {
		return
	}
	pth, err := self.findPath(from)
	if err != nil {
		return
	}

	self.root = self.obj.AddLink(self.root, pth, to)
	self.root = self.obj.RmLink(self.root, from)
	if isPathContained(from, self.nowPathStr) {
		self.nowPathStr = toAbsPath(from, "..")
	} else {
		self.nowPathStr = toAbsPath(self.nowPathStr, self.nowPathStr)
	}
	self.nowPath, _ = self.findPath(self.nowPathStr)
}
func (self *fileSystem) Rm(from string) {
	from = toAbsPath(self.nowPathStr, from)
	if from == "" {
		return
	}
	_, err := self.findPath(from)
	if err != nil {
		return
	}

	self.root = self.obj.RmLink(self.root, from)
	if isPathContained(from, self.nowPathStr) {
		self.nowPathStr = toAbsPath(from, "..")
	} else {
		self.nowPathStr = toAbsPath(self.nowPathStr, self.nowPathStr)
	}
	self.nowPath, _ = self.findPath(self.nowPathStr)
}
func (self *fileSystem) Ls() []link {
	if ok := isLinked(self.nowPath, self.obj); !ok {
		return nil
	}
	return self.obj.Links(self.nowPath)
}
func (self *fileSystem) Mldir(to string) {
	to = toAbsPath(self.nowPathStr, to)
	self.root = self.obj.AddLink(self.root, newEmptyDir(), to)
	self.nowPathStr = toAbsPath(self.nowPathStr, self.nowPathStr)
	self.nowPath, _ = self.findPath(self.nowPathStr)
}
func (self *fileSystem) findPath(to string) (ipath.Resolved, error) {
	if to == "" {
		return self.root, nil
	}

	pth := self.root
	pathNameList := strings.Split(to, "/")
	for _, name := range pathNameList {
		if ok := isLinked(pth, self.obj); !ok {
			return nil, util.NewError("the path is not linked")
		}
		if ok := isDir(pth, self.obj); !ok {
			return nil, util.NewError("the path is not a directory")
		}

		isMatched := false
		for _, link := range self.obj.Links(pth) {
			if link.Name == name {
				pth = link.Path
				isMatched = true
				break
			}
		}
		if !isMatched {
			return nil, util.NewError("the path does not exist")
		}
	}
	return pth, nil
}

func isLinked(pth ipath.Path, obj *object) bool {
	stat, err := obj.Stat(pth)
	if err != nil {
		return false
	}
	return stat.NumLinks > 0
}
func isDir(pth ipath.Path, obj *object) bool {
	nd, _ := obj.Get(pth)
	pbnd := nd.(*dag.ProtoNode)
	fsnd, _ := unixfs.FSNodeFromBytes(pbnd.Data())
	return fsnd.IsDir()
}
func isPathContained(upper, lower string) bool {
	upList := strings.Split(upper, "/")
	lowList := strings.Split(lower, "/")
	if len(upList) > len(lowList) {
		return false
	}

	for idx, _ := range upList {
		if upList[idx] != lowList[idx] {
			return false
		}
	}
	return true
}

func toAbsPath(from, to string) string {
	if strings.HasPrefix(to, "~") {
		to = strings.TrimLeft(to, "~")
	}
	if filepath.IsAbs(to) {
		return strings.TrimLeft(to, "/")
	}
	if from == to {
		return from
	}
	if ok := filepath.IsAbs(from); !ok {
		from = "/" + from
	}
	pth := filepath.Join(from, to)
	return strings.TrimLeft(pth, "/")
}

func newEmptyDir() ipath.Path {
	pbnd := unixfs.EmptyDirNode()
	return ipath.IpfsPath(pbnd.Cid())
}
