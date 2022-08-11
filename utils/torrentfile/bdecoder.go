package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

const ()

// 文件数据信息
type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`       // 字节序列
	PieceLength int    `bencode:"piece length"` // 分片长度
	Length      int    `bencode:"length"`       // 文件大小 以字节为单位
	Name        string `bencode:"name"`         // 资源名称
}

// 生成info_hash
func (bi *bencodeInfo) hash() ([20]byte, error) {
	//todo: 根据meta info 生成 infohash

	var buf bytes.Buffer

	err := bencode.Marshal(&buf, *bi)

	if err != nil {
		return [20]byte{}, err
	}

	// 根据info信息生成sha1校验值
	infoHash := sha1.Sum(buf.Bytes())

	return infoHash, nil

}

// 生成pieceHashes
func (bi *bencodeInfo) splitePieces() ([][20]byte, error) {
	// todo: 根据.torrent文件切分pieces

	// 每个piece的长度位20
	hashLen := 20

	piecesBuf := []byte(bi.Pieces)

	// piece的数量等于pieces的总长度/piece的长度
	totalPieceLen := len(piecesBuf) / hashLen

	if totalPieceLen != 0 {
		return [][20]byte{}, fmt.Errorf("torrent file meta info: [pieces] is malformed")
	}

	pieceHashes := make([][20]byte, totalPieceLen)

	for i := 0; i < totalPieceLen; i++ {
		copy(pieceHashes[i][:], piecesBuf[i*hashLen:(i+1)*hashLen])
	}

	return pieceHashes, nil
}

// 元数据信息
type bencodeTorrent struct {
	Announce string      `bencode:"announce"` // tacker 地址
	Info     bencodeInfo `bencode:"info"`
}

func (bto *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	//todo: 根据.torrent文件生成torrent对象

	announce := bto.Announce

	infohash, err := bto.Info.hash()

	if err != nil {
		return TorrentFile{}, err
	}

	pieceHashes, err := bto.Info.splitePieces()

	if err != nil {
		return TorrentFile{}, err
	}

	tf := TorrentFile{
		Announce:    announce,
		InfoHash:    infohash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}

	return tf, nil
}

// 读取.torrent种子文件,生成torrentFile对象
func Open(path string) (TorrentFile, error) {
	//todo: 读取.torrent 文件生成torrentFile文件

	r, err := os.OpenFile(path, os.O_RDONLY, 0777)

	if err != nil {
		return TorrentFile{}, err
	}

	bto := bencodeTorrent{}

	err = bencode.Unmarshal(r, &bto)

	if err != nil {
		return TorrentFile{}, err
	}

	return bto.toTorrentFile()
}
