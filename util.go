package ipfs

import (
	"errors"
	"os"

	files "github.com/ipfs/go-ipfs-files"

	"github.com/pilinsin/util"
)

func bytesToIpfsFile(b []byte) files.File {
	return files.NewBytesFile(b)
}

func ipfsFileNodeToBytes(node files.Node) ([]byte, error) {
	switch node := node.(type) {
	case files.File:
		return util.ReaderToBytes(node), nil
	default:
		err := errors.New("node (type: files.Node) does not have files.File type!")
		return nil, err
	}
}
func filePathToIpfsFile(filePath string) (files.File, error) {
	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		return nil, err
	} else {
		return files.NewReaderFile(file), nil
	}
}

func dirPathToIpfsFileNode(dirPath string) (files.Node, error) {
	st, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}
	return files.NewSerialFile(dirPath, false, st)
}
