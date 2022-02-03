package ipfs

import (
	"context"
	"io/ioutil"
	"os"

	kstore "github.com/ipfs/go-ipfs-keystore"
	core "github.com/ipfs/go-ipfs/core"
	coreapi "github.com/ipfs/go-ipfs/core/coreapi"
	libp2p "github.com/ipfs/go-ipfs/core/node/libp2p"
	iface "github.com/ipfs/interface-go-ipfs-core"

	"github.com/pilinsin/util"
)

type IPFS struct {
	api    iface.CoreAPI
	ctx    context.Context
	kStore kstore.Keystore
	node *core.IpfsNode
	close func()
}
func New(mode string) (*IPFS, error) {
	repoPath, _ := ioutil.TempDir("", "ipfs-tmp")
	switch mode {
	case "tor":
		return newIpfsTor(repoPath)
	default:
		return newIpfs(repoPath)
	}
}
func newIpfs(repoPath string) (*IPFS, error){
	r, err := newRepo(repoPath)
	if err != nil {
		return nil, err
	}
	exOpts := map[string]bool{
		"discovery": true,
		"dht":       true,
		"pubsub":    true,
	}
	ctx := context.Background()
	buildCfg := &core.BuildCfg{
		Online:    true,
		Repo:      r,
		Routing:   libp2p.DHTClientOption,
		ExtraOpts: exOpts,
	}
	node, _ := core.NewNode(ctx, buildCfg)
	api, _ := coreapi.NewCoreAPI(node)
	close := func(){
		node.Close()
		os.RemoveAll(repoPath)
	}
	return &IPFS{api, ctx, r.Keystore(), node, close}, nil
}
func newIpfsTor(repoPath string) (*IPFS, error){
	return nil, util.NewError("unimplemented error")
/*
	swarmAddr, hostOpt, dnsResolver, err := newTor(repoPath)
	if err != nil{return nil, err}

	r, err := newRepo(repoPath)
	if err != nil {
		return nil, err
	}
	exOpts := map[string]bool{
		"discovery": true,
		"dht":       true,
		"pubsub":    true,
	}
	rCfg, _ := r.Config()
	rCfg.Addresses.Swarm = []string{swarmAddr}

	ctx := context.Background()
	buildCfg := &core.BuildCfg{
		Online:    true,
		Repo:      r,
		Host: hostOpt,
		Routing:   libp2p.DHTOption,
		ExtraOpts: exOpts,
	}
	node, _ := core.NewNode(ctx, buildCfg)
	node.DNSResolver = dnsResolver
	api, _ := coreapi.NewCoreAPI(node)
	close := func(){
		node.Close()
		os.RemoveAll(repoPath)
	}
	return &IPFS{api, ctx, r.Keystore(), node, close}, nil
*/
}
func (ipfs *IPFS) Close() {
	ipfs.api = nil
	ipfs.close()
}
