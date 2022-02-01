package hls

import(
	"os"
	"bufio"
	"path/filepath"

	util "github.com/pilinsin/util"
	m3u8 "github.com/grafov/m3u8"
	ipfs "github.com/pilinsin/ipfs-util"
)

func addDirPath(parent, child string) string{
	dir, _ := filepath.Split(parent)
	return filepath.Join(dir, child)
}

func ConvertAndAdd(input, fmpgPath, fprbPath, httpGatewayIpfs string, is *ipfs.IPFS) (string, error){
	output := addDirPath(input, "output_video")
	m3u8Path, err := util.VideoToHLS(input, output, fmpgPath, fprbPath)
	if err != nil{return "", err}
	return Add(m3u8Path, httpGatewayIpfs, is)
}

func Add(m3u8Path, httpGatewayIpfs string, is *ipfs.IPFS) (string, error){
	f, err := os.Open(m3u8Path)
	if err != nil{return "", err}
	defer f.Close()
	p, listType, err := m3u8.DecodeFrom(bufio.NewReader(f), true)
	if err != nil{return "", err}

	switch listType {
	case m3u8.MEDIA:
		playList := p.(*m3u8.MediaPlaylist)
		for _, segment := range playList.Segments{
			if segment == nil{break}
			uri := addDirPath(m3u8Path, segment.URI)
			data, err := os.ReadFile(uri)
			if err != nil{return "", err}
			cid := ipfs.File.Add(data, is)
			segment.URI = httpGatewayIpfs + cid
			//os.Remove(uri)
		}
	default:
		playList := p.(*m3u8.MasterPlaylist)
		for _, variant := range playList.Variants{
			if variant == nil{break}
			uri := addDirPath(m3u8Path, variant.URI)
			data, err := os.ReadFile(uri)
			if err != nil{return "", err}
			cid := ipfs.File.Add(data, is)
			variant.URI = httpGatewayIpfs + cid

			for _, segment := range variant.Chunklist.Segments{
				if segment == nil{break}
				uri := addDirPath(m3u8Path, segment.URI)
				data, err := os.ReadFile(uri)
				if err != nil{return "", err}
				cid := ipfs.File.Add(data, is)
				segment.URI = httpGatewayIpfs + cid
				//os.Remove(uri)
			}
			//os.Remove(uri)
		}

	}
	m := p.Encode().Bytes()
	os.WriteFile(addDirPath(m3u8Path, "tmp.m3u8"), m, 0755)
	return ipfs.File.Add(m, is), nil
}