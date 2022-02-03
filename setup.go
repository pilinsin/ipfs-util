package ipfs

import (
	"io/ioutil"
	"os"
	"path/filepath"

/*
	"context"
	"crypto/ed25519"

	tor "github.com/cretz/bine/tor"
	libtor "github.com/ipsn/go-libtor"
	madns "github.com/multiformats/go-multiaddr-dns"
	oniontp "github.com/cpacia/go-onion-transport"
	libp2p "github.com/libp2p/go-libp2p"
	p2pconf "github.com/libp2p/go-libp2p/config"
	host "github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
*/

	config "github.com/ipfs/go-ipfs-config"
	assets "github.com/ipfs/go-ipfs/assets"
	core "github.com/ipfs/go-ipfs/core"
	plugin "github.com/ipfs/go-ipfs/plugin"
	flatfs "github.com/ipfs/go-ipfs/plugin/plugins/flatfs"
	levelds "github.com/ipfs/go-ipfs/plugin/plugins/levelds"
	repo "github.com/ipfs/go-ipfs/repo"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	"github.com/ipfs/go-ipfs/repo/fsrepo/migrations"
	"github.com/ipfs/go-ipfs/repo/fsrepo/migrations/ipfsfetcher"
	path "github.com/ipfs/go-path"
	unixfs "github.com/ipfs/go-unixfs"
	options "github.com/ipfs/interface-go-ipfs-core/options"

	"github.com/pilinsin/util"
)
/*
func newTor(repoPath string) (string, libp2p.HostOption, *madns.Resolver, error){
	torDir := filepath.Join(repoPath, "tor")
	client, err := tor.Start(nil, &tor.StartConf{
		ProcessCreator: libtor.Creator,
		DataDir: torDir,
		NoAutoSocksPort: true,
		EnableNetwork: true,
		ExtraArgs: []string{"--DNSPort", "2121"},
	})
	if err != nil{return "", nil, nil, err}

	dialer, err := client.Dialer(util.NewContext(), nil)
	if err != nil{return "", nil, nil, err}
	_, priv, _ := ed25519.GenerateKey(nil)
	service, err := client.Listen(util.NewContext(), &tor.ListenConf{
		RemotePorts: []int{9003},
		Version3: true,
		Key: priv,
	})
	if err != nil{return "", nil, nil, err}

	madns.DefaultResolver = oniontp.NewTorResolver("localhost:2121")
	dialOnionOnly := true
	tpOpt := libp2p.Transport(oniontp.NewOnionTransportC(dialer, service, dialOnionOnly))
	torHost := func(ctx context.Context, id peer.ID, ps pstore.PeerStore, opts ...libp2p.Option) (host.Host, err){
		priKey := ps.PrivKey(id)
		if priv == nil{return nil, fmt.Errorf("missing private key for node ID: %s", id,Pretty())}
		opts = append([]libp2p.Option{libp2p.Identity(priKey), libp2p.Peerstore(ps)}, opts...)
		cfg := &p2pconf.config{}
		if err := cfg.Apply(opts...); err != nil{return nil, err}
		cfg.Transports = nil
		if err := tpOpt(cfg); err != nil{return nil, err}
		return cfg.NewNode(ctx)
	}
	swarmAddr := fmt.Sprintf("/onion3/%s:9003", service.ID)

	return swarmAddr, torHost, madns.DefaultResolver, nil
}
*/

func newRepo(repoPath string) (repo.Repo, error) {
	//ipfsKeyInit()
	cfg, err := newConfig("")
	if err != nil {
		return nil, err
	}
	loadPlugins()
	if err := doInit(repoPath, cfg); err != nil {
		return nil, err
	}
	//if err := migration(repoPath); err != nil{return nil, err}

	return fsrepo.Open(repoPath)
}

func loadPlugins() {
	//ignore "already have a datastore named ~" errors
	fp := flatfs.Plugins[0].(plugin.PluginDatastore)
	fsrepo.AddDatastoreConfigHandler("flatfs", fp.DatastoreConfigParser())
	lp := levelds.Plugins[0].(plugin.PluginDatastore)
	fsrepo.AddDatastoreConfigHandler("levelds", lp.DatastoreConfigParser())
}

func newConfig(mode string) (*config.Config, error) {
	var id config.Identity
	var err error
	switch mode {
	case "original":
		id, err = newOriginalKeydentity()
	default:
		keyGenOpts := []options.KeyGenerateOption{options.Key.Type(options.Ed25519Key)}
		id, err = config.CreateIdentity(ioutil.Discard, keyGenOpts)
	}
	if err != nil {
		return nil, err
	}
	return config.InitWithIdentity(id)
}

func doInit(repoPath string, cfg *config.Config) error {
	if err := checkWritable(repoPath); err != nil {
		return err
	}
	if fsrepo.IsInitialized(repoPath) {
		return util.NewError("ipfs config file already exists.")
	}
	if err := fsrepo.Init(repoPath, cfg); err != nil {
		return err
	}
	if err := addDefaultAssets(repoPath); err != nil {
		return err
	}

	return initializeIpnsKeyspace(repoPath)
}

func checkWritable(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(dir, 0775)
		}
		if os.IsPermission(err) {
			return util.NewError("unexpected error:", err)
		}
		return err
	}
	testFile := filepath.Join(dir, "test")
	f, err := os.Create(testFile)
	if err != nil {
		if os.IsPermission(err) {
			return util.NewError(dir, "is not writable by the user")
		}
		return util.NewError("unexpected error:", err)
	}
	f.Close()
	return os.Remove(testFile)
}
func addDefaultAssets(repoPath string) error {
	ctx, cancel := util.CancelContext()
	defer cancel()

	r, err := fsrepo.Open(repoPath)
	if err != nil {
		return err
	}
	node, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer node.Close()
	_, err = assets.SeedInitDocs(node)
	if err != nil {
		return util.NewError("seeding init docs failed:", err)
	}

	return nil
}
func initializeIpnsKeyspace(repoPath string) error {
	ctx, cancel := util.CancelContext()
	defer cancel()

	r, err := fsrepo.Open(repoPath)
	if err != nil {
		return err
	}
	node, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer node.Close()

	emptyDir := unixfs.EmptyDirNode()
	if err := node.Pinning.Pin(ctx, emptyDir, true); err != nil {
		return err
	}
	if err := node.Pinning.Flush(ctx); err != nil {
		return err
	}

	return node.Namesys.Publish(ctx, node.PrivateKey, path.FromCid(emptyDir.Cid()))
}

func migration(repoPath string) error {
	migrationCfg, err := migrations.ReadMigrationConfig(repoPath)
	if err != nil {
		return err
	}

	fetcherFunc := func(distPath string) migrations.Fetcher {
		return ipfsfetcher.NewIpfsFetcher(distPath, 0, &repoPath)
	}
	fetchDistPath := migrations.GetDistPathEnv(migrations.CurrentIpfsDist)
	fetcher, err := migrations.GetMigrationFetcher(migrationCfg.DownloadSources, fetchDistPath, fetcherFunc)
	if err != nil {
		return err
	}
	defer fetcher.Close()

	cache := migrationCfg.Keep == "cache"
	pin := migrationCfg.Keep == "pin"
	if cache || pin {
		migrations.DownloadDirectory, err = ioutil.TempDir("", "migrations")
		if err != nil {
			return err
		}
		defer func() {
			if migrations.DownloadDirectory != "" {
				os.RemoveAll(migrations.DownloadDirectory)
			}
		}()
	}
	return migrations.RunMigration(util.NewContext(), fetcher, fsrepo.RepoVersion, "", false)
}
