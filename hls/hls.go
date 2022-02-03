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

func ConvertAndAdd(input, fmpgPath, fprbPath string, is *ipfs.IPFS) (string, error){
	output := addDirPath(input, "output_video")
	m3u8Path, err := util.VideoToHLS(input, output, fmpgPath, fprbPath)
	if err != nil{return "", err}
	return Add(m3u8Path, is)
}

func Add(m3u8Path string, is *ipfs.IPFS) (string, error){
	f, err := os.Open(m3u8Path)
	if err != nil{return "", err}
	defer f.Close()
	p, listType, err := m3u8.DecodeFrom(bufio.NewReader(f), true)
	if err != nil{return "", err}

	ipfsFs := ipfs.Object.NewFileSystem(is)

	switch listType {
	case m3u8.MEDIA:
		playList := p.(*m3u8.MediaPlaylist)
		for _, segment := range playList.Segments{
			if segment == nil{break}
			uri := addDirPath(m3u8Path, segment.URI)
			data, err := os.ReadFile(uri)
			if err != nil{return "", err}
			pth := is.File().Add(data, true)
			ipfsFs.Add(pth, segment.URI)
			//os.Remove(segment.URI)
		}
	default:
		playList := p.(*m3u8.MasterPlaylist)
		for _, variant := range playList.Variants{
			if variant == nil{break}
			uri := addDirPath(m3u8Path, variant.URI)
			data, err := os.ReadFile(uri)
			if err != nil{return "", err}
			pth := is.File().Add(data, true)
			ipfsFs.Add(pth, variant.URI)

			for _, segment := range variant.Chunklist.Segments{
				if segment == nil{break}
				uri := addDirPath(m3u8Path, segment.URI)
				data, err := os.ReadFile(uri)
				if err != nil{return "", err}
				pth := is.File().Add(data, true)
				ipfsFs.Add(pth, segment.URI)
				//os.Remove(segment.URI)
			}
			//os.Remove(variant.URI)
		}

	}
	m := p.Encode().Bytes()
	pth := is.File().Add(m, true)
	_, m3u8FileName := filepath.Split(m3u8Path)
	ipfsFs.Add(pth, m3u8FileName)
	return pth.Cid().String(), nil
}