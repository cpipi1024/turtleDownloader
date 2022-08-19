package torrentfile

import (
	"crypto/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"cpipi1024.com/turtleDownloader/utils/downloader"
	"cpipi1024.com/turtleDownloader/utils/peers"
	"github.com/jackpal/bencode-go"
)

const (
	Port = 6881
)

// BT tracker响应对象
type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (t *TorrentFile) DownLoad(path string) error {
	var peerId [20]byte

	_, err := rand.Read(peerId[:])

	if err != nil {
		return err
	}

	peers, err := t.requestPeers(peerId, Port)
	if err != nil {
		return err
	}

	torrent := &downloader.Torrent{
		Peers:       peers,
		PeerID:      peerId,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}

	buf, err := torrent.Download()

	if err != nil {
		return nil
	}

	outfile, err := os.Create(path)

	if err != nil {
		return err
	}

	defer outfile.Close()
	_, err = outfile.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

// 构建tracker地址
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

// 向tracker请求peers
func (t *TorrentFile) requestPeers(peerID [20]byte, port uint) ([]peers.Peer, error) {
	url, err := t.builTrackerURL(peerID, port)

	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	trackerResp := bencodeTrackerResp{}

	err = bencode.Unmarshal(resp.Body, &trackerResp)

	if err != nil {
		return nil, err
	}

	peersBytes := []byte(trackerResp.Peers)

	return peers.Unmarshal(peersBytes)
}
