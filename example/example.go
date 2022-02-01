package main

import (
	"fmt"
	"time"
	"github.com/pilinsin/ipfs-util"
	hls "github.com/pilinsin/ipfs-util/hls"
	scmap "github.com/pilinsin/ipfs-util/scalablemap"
)

func main() {
	is, err := ipfs.New(false)
	if err != nil{
		fmt.Println(err)
		return
	}
	//hlsExample(is)
//	/*
	data := []byte("meow meow 3 ^.^")
	fs := ipfs.Object.NewFileSystem(is)
	pth := is.File().Add(data, true)
	fmt.Println(pth)
	data0, err := is.File().Get(pth)
	fmt.Println(string(data0), err)
	fs.Add(pth, "data.txt")
	fmt.Println(fs.Root())
	pth0, err := fs.Get("/data.txt")
	fmt.Println(pth0, err)
//	*/

//	/*
	mapExample("const", is)
	mapExample("ordered", is)

	fileExample(is)
	nameExample(is)
//	*/
//	/*
	objectExample(is)

	is2, err := ipfs.New(false)
	if err != nil{
		fmt.Println(err)
		return
	}
	pubsubExample(is, is2)
//	*/
}
func hlsExample(is *ipfs.IPFS){
	cid, err := hls.ConvertAndAdd("/home/yo-cana/test/kyuuketsuki 01.mp4", "/usr/bin/ffmpeg", "/usr/bin/ffprobe", "https://ipfs.io/ipfs/", is)
	fmt.Println(cid, err)	
}

func fileExample(is *ipfs.IPFS) {
	data := []byte("meow meow ^.^")

	cid := ipfs.File.Add(data, is)
	data1, _ := ipfs.File.Get(cid, is)
	fmt.Println(data)
	fmt.Println(data1)
}
func nameExample(is *ipfs.IPFS) {
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
func pubsubExample(is, is2 *ipfs.IPFS){
	sub := is.PubSub().Subscribe("test_topic")
	defer sub.Close()
	<-time.Tick(10*time.Second)
	N := 2
	go func(){
		for i := 0; i < N; i++{
			is2.PubSub().Publish([]byte(fmt.Sprintf("message: %3d", i)), "test_topic")
		}
	}()
	n := 0
	for{
		mesList := is.PubSub().NextAll(sub)
		fmt.Println(len(mesList))
		for idx, mes := range mesList{
			fmt.Println(idx, string(mes))
		}
		n += len(mesList)
		if n >= N{return}
		<-time.Tick(1*time.Second)
	}
}
func objectExample(is *ipfs.IPFS){
	fs := ipfs.Object.NewFileSystem(is)
	data := is.File().Add([]byte("meow meow ^.^"), true)
	fs.Add(data, "file0.dat")
	fmt.Println(fs.Ls())
	fs.Mkdir("a")
	fs.Mkdir("b")
	fs.Mkdir("c")
	fmt.Println(fs.Ls())
	fs.Cp("file0.dat", "a/file0cp.dat")
	fs.Cd("a")
	fmt.Println(fs.Ls())
	fs.Mv("/file0.dat", "/b/b1/b2/b3/file0mv.dat")
	fs.Cd("/b/b1/b2/b3")
	fmt.Println(fs.Ls())
	fs.Rm("../..")
	fmt.Println(fs.Ls())
	fs.Cp("/a/file0cp.dat", "/file02.dat")
	fs.Cd("/")
	fmt.Println(fs.Ls())
	fmt.Println(fs.GetAllFiles(fs.Root()))
}

func mapExample(mode string, is *ipfs.IPFS) {
	vm := scmap.NewScalableMap(mode, 10000)
	vm.Append("a", nil, is)
	m := vm.Marshal()
	vm2, err := scmap.UnmarshalScalableMap(mode, m)
	cid := ipfs.File.Add(m, is)
	fmt.Println(vm2.ContainCid(cid, is), err)
}
