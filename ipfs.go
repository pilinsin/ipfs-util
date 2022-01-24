package ipfs

import (
	"context"
	"io/ioutil"

	kstore "github.com/ipfs/go-ipfs-keystore"
	core "github.com/ipfs/go-ipfs/core"
	coreapi "github.com/ipfs/go-ipfs/core/coreapi"
	libp2p "github.com/ipfs/go-ipfs/core/node/libp2p"
	iface "github.com/ipfs/interface-go-ipfs-core"
)

type IPFS struct {
	api iface.CoreAPI
	ctx     context.Context
	kStore  kstore.Keystore
}

func New(repoStr string) (*IPFS, error) {
	repoPath, _ := ioutil.TempDir("", repoStr)
	r, err := newRepo(repoPath)
	if err != nil {
		return nil, err
	}

	exOpts := map[string]bool{
		"discovery": true,
		"dht":       true,
		"pubsub":    true,
	}
	buildCfg := core.BuildCfg{
		Online:    true,
		Repo:      r,
		Routing: libp2p.DHTOption,
		ExtraOpts: exOpts,
	}

	ctx := context.Background()
	node, _ := core.NewNode(ctx, &buildCfg)
	api, _ := coreapi.NewCoreAPI(node)

	return &IPFS{api, ctx, r.Keystore()}, nil
}
func (ipfs *IPFS) Close() {
	ipfs.api = nil
}
