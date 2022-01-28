package main

import(
	"fmt"
	"github.com/pilinsin/ipfs-util"
	scmap "github.com/pilinsin/ipfs-util/scalablemap"
)

func main(){
	is, _ := ipfs.New("test-ipfs")
	mapExample("const", is)
	mapExample("ordered", is)

	fileExample(is)
	nameExample(is)	
}

func fileExample(is *ipfs.IPFS){
	data := []byte("meow meow ^.^")

	cid := ipfs.File.Add(data, is)
	data1, _ := ipfs.File.Get(cid, is)
	fmt.Println(data)
	fmt.Println(data1)
}
func nameExample(is *ipfs.IPFS){
	data := []byte("meow meow ^.^")
	cid := ipfs.File.Hash(data, is)

	kf := ipfs.Name.NewKeyFile()
	kf2 := &ipfs.KeyFile{}
	kf2.Unmarshal(kf.Marshal())
	fmt.Println(kf.Equals(kf2))
	name, _ := kf.Name()
	name1 := ipfs.Name.PublishWithKeyFile(data, kf, is)
	name2 := ipfs.Name.PublishCidWithKeyFile(cid, kf, is)
	fmt.Println(name)
	fmt.Println(name1)
	fmt.Println(name2)
	data2, _ := ipfs.Name.Get(name, is)
	fmt.Println(data2)
	cid1, _ := ipfs.Name.GetCid(name, is)
	fmt.Println(cid)
	fmt.Println(cid1)
}


func mapExample(mode string, is *ipfs.IPFS){
	vm := scmap.NewScalableMap(mode, 10000)
	vm.Append("a", nil, is)
	m := vm.Marshal()
	vm2, err := scmap.UnmarshalScalableMap(mode, m)
	cid := ipfs.File.Add(m, is)
	fmt.Println(vm2.ContainCid(cid, is), err)
}