package ipfs

import (
	"context"

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
}

func New(tor bool) (*IPFS, error) {
	if tor{
		return newIpfsWithTor()
	}else{
		return newIpfs()
	}
}
func newIpfs() (*IPFS, error){
	r, err := newRepo()
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
		Routing:   libp2p.DHTOption,
		ExtraOpts: exOpts,
	}
	node, _ := core.NewNode(ctx, buildCfg)
	api, _ := coreapi.NewCoreAPI(node)
	return &IPFS{api, ctx, r.Keystore()}, nil
}
func newIpfsWithTor() (*IPFS, error){
	return nil, util.NewError("unimplemented error")
/*
	swarmAddr, hostOpt, dnsResolver, err := newTor()
	if err != nil{return nil, err}

	r, err := newRepo()
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
	return &IPFS{api, ctx, r.Keystore()}, nil
*/
}
func (ipfs *IPFS) Close() {
	ipfs.api = nil
}
