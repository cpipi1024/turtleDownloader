package utils

import (
	"io"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

// 文件数据信息
type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`       // 字节序列
	PieceLength int    `bencode:"piece length"` // 分片长度
	Length      int    `bencode:"length"`       // 文件大小 以字节为单位
	Name        string `bencode:"name"`         // 资源名称
}

// 元数据信息
type bencodeTorrent struct {
	Announce string      `bencode:"announce` // tacker 地址
	Info     bencodeInfo `bencode:"info"`
}

func Open(r io.Reader) (*bencodeTorrent, error) {
	//todo: parse bencode data
	bto := bencodeTorrent{}

	err := bencode.Unmarshal(r, &bto)

	if err != nil {
		return nil, err
	}

	return &bto, nil
}

//
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (bto bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	//todo:根据.torrent文件生成torrent对象
}

func (t *TorrentFile) builTrackerURL(peerID [20]byte, port uint) (string, error) {
	//todo: 根据 announce生成bt trakcer请求

	// announce "http:xxxbttracker.com:port/source"
	base, err := url.Parse(t.Announce)

	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}

	base.RawQuery = params.Encode()

	return base.String(), nil

}
