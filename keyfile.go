package ipfs

import (
	iface "github.com/ipfs/interface-go-ipfs-core"
	p2pcrypt "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

var KeyFileMode  = "ed25519"

type KeyFile struct {
	keyFile p2pcrypt.PrivKey
}
func (self nameUtil) NewKeyFile() *KeyFile {
	var priv p2pcrypt.PrivKey
	switch KeyFileMode {
	case "ed25519":
		priv, _, _ = p2pcrypt.GenerateKeyPair(p2pcrypt.Ed25519, -1)
	default:
		priv, _ = newOriginalPrivKeyPair()
	}
	return &KeyFile{priv}
}
func (kf *KeyFile) Name() (string, error) {
	pid, err := peer.IDFromPrivateKey(kf.keyFile)
	if err != nil {
		return "", err
	} else {
		name := iface.FormatKeyID(pid)
		return name, nil
	}
}
func (kf *KeyFile) Equals(kf2 *KeyFile) bool {
	return kf.keyFile.Equals(kf2.keyFile)
}
func (kf KeyFile) Marshal() []byte {
	kb, _ := p2pcrypt.MarshalPrivateKey(kf.keyFile)
	return kb
}
func (kf *KeyFile) Unmarshal(b []byte) error {
	kFile, err := p2pcrypt.UnmarshalPrivateKey(b)
	if err == nil {
		kf.keyFile = kFile
		return nil
	} else {
		return err
	}
}
