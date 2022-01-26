package ipfs

import(
	//"fmt"
	"encoding/base64"

	config "github.com/ipfs/go-ipfs-config"
	peer "github.com/libp2p/go-libp2p-core/peer"
	p2pcrypt "github.com/libp2p/go-libp2p-core/crypto"
	pb "github.com/libp2p/go-libp2p-core/crypto/pb"
	"github.com/pilinsin/util"
	"github.com/pilinsin/util/crypto"
)

const KeyType_Original pb.KeyType = 100

func ipfsKeyInit(){
	p2pcrypt.KeyTypes = append(p2pcrypt.KeyTypes, 100)
	pb.KeyType_name[100] = "Original"
	pb.KeyType_value["Original"] = 100

	p2pcrypt.PubKeyUnmarshallers[KeyType_Original] = UnmarshalOriginalPubKey
	p2pcrypt.PrivKeyUnmarshallers[KeyType_Original] = UnmarshalOriginalPrivKey
}
func newOriginalKeydentity() (config.Identity, error){
	id := config.Identity{}
	sk, pk := newOriginalPrivKeyPair()
	
	msk, err := p2pcrypt.MarshalPrivateKey(sk)
	if err != nil{return id, err}
	id.PrivKey = base64.StdEncoding.EncodeToString(msk)

	pid, err := peer.IDFromPublicKey(pk)
	if err != nil{return id, err}
	id.PeerID = pid.Pretty()

	return id, nil
}

type originalPrivKey struct{
	priKey crypto.ISignKey
	pubKey crypto.IVerfKey
}
func newOriginalPrivKeyPair() (p2pcrypt.PrivKey, p2pcrypt.PubKey){
	kp := crypto.NewSignKeyPair()
	return &originalPrivKey{kp.Sign(), kp.Verify()}, &originalPubKey{kp.Verify()}
}
func (pri *originalPrivKey) Equals(pri2 p2pcrypt.Key) bool{
	m1, err := pri.Raw()
	if err != nil{return false}
	m2, err := pri2.Raw()
	if err != nil{return false}
	return util.ConstTimeBytesEqual(m1, m2)
}
func (pri *originalPrivKey) Raw() ([]byte, error){
	mpri := &struct{Pr, Pu []byte}{pri.priKey.Marshal(), pri.pubKey.Marshal()}
	return util.Marshal(mpri)
}
func (pri *originalPrivKey) Type() pb.KeyType{
	return KeyType_Original
}
func (pri *originalPrivKey) Sign(data []byte) ([]byte, error){
	return pri.priKey.Sign(data)
}
func (pri *originalPrivKey) GetPublic() p2pcrypt.PubKey{
	return &originalPubKey{pri.pubKey}
}
func UnmarshalOriginalPrivKey(m []byte) (p2pcrypt.PrivKey, error){
	mpri := &struct{Pr, Pu []byte}{}
	if err := util.Unmarshal(m, mpri); err != nil{return nil, err}
	sk, err := crypto.UnmarshalSignKey(mpri.Pr)
	if err != nil{return nil, err}
	vk, err := crypto.UnmarshalVerfKey(mpri.Pu)
	if err != nil{return nil, err}

	pri := &originalPrivKey{}
	pri.priKey = sk
	pri.pubKey = vk
	return pri, nil
}

type originalPubKey struct{
	pubKey crypto.IVerfKey
}
func (pub *originalPubKey) Equals(pub2 p2pcrypt.Key) bool{
	m1, err := pub.Raw()
	if err != nil{return false}
	m2, err := pub2.Raw()
	if err != nil{return false}
	return util.ConstTimeBytesEqual(m1, m2)
}
func (pub *originalPubKey) Raw() ([]byte, error){
	return pub.pubKey.Marshal(), nil
}
func (pub *originalPubKey) Type() pb.KeyType{
	return KeyType_Original
}
func (pub *originalPubKey) Verify(data, sig []byte) (bool, error){
	return pub.pubKey.Verify(data, sig)
}
func UnmarshalOriginalPubKey(m []byte) (p2pcrypt.PubKey, error){
	vk, err := crypto.UnmarshalVerfKey(m)
	return &originalPubKey{vk}, err
}