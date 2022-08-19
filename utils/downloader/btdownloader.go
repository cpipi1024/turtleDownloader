package downloader

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"time"

	"cpipi1024.com/turtleDownloader/client"
	"cpipi1024.com/turtleDownloader/utils/message"
	"cpipi1024.com/turtleDownloader/utils/peers"
)

const (
	MaxBacklogSize = 16384
	MaXBacklog     = 5
)

// torrent 保存远端peers和本地peer端信息
type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()

	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		index, err := message.ParseMsgHave(msg)
		if err != nil {
			return err
		}
		state.client.BitField.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParseMsgPiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func attempDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	// 开始下载
	for state.downloaded < pw.length {
		if !state.client.Choked {
			for state.backlog < MaXBacklog && state.requested < pw.index {
				blockSize := MaxBacklogSize

				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)

				if err != nil {
					return nil, err
				}

				state.backlog++
				state.requested += blockSize
			}
		}
		err := state.readMessage()

		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)

	if bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("index %d failed intergrity check ", pw.index)
	}

	return nil
}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("start download for ", t.Name)

	workQueue := make(chan *pieceWork, len(t.PieceHashes))

	results := make(chan *pieceResult)

	for idx, hash := range t.PieceHashes {
		length := t.calculatePieceSize(idx)
		workQueue <- &pieceWork{idx, hash, length}
	}

	// 启动woker
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	buf := make([]byte, t.Length)

	doncePieces := 0

	for doncePieces <= len(t.PieceHashes) {
		res := <-results

		begin, end := t.calculateBoundsForPiece(res.index)

		copy(buf[begin:end], res.buf)

		doncePieces++

		percents := float64(doncePieces) / float64(len(t.PieceHashes)) * 100

		numWorkers := runtime.NumGoroutine()

		log.Printf("(%0.2f%%) downloaded piece #%d from #%dpeers\n", percents, res.index, numWorkers)

	}

	close(workQueue)

	return buf, nil

}

func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := client.NewClient(peer, t.PeerID, t.InfoHash)

	if err != nil {
		log.Println("could not handshake with:", peer.IP)
		return
	}

	defer c.Conn.Close()

	log.Printf("completed handshake with %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQueue {
		if !c.BitField.HasPiece(pw.index) {
			workQueue <- pw
			continue
		}

		buf, err := attempDownloadPiece(c, pw)

		if err != nil {
			log.Panicln("Exiting:", err)
			workQueue <- pw
			return
		}

		err = checkIntegrity(pw, buf)

		if err != nil {
			log.Printf("piece #%d failed check integrity check \n", pw.index)
			workQueue <- pw
			continue
		}
		c.SendHave(pw.index)

		results <- &pieceResult{pw.index, buf}
	}
}

// 返回下载的piece大小
func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)

	return end - begin
}

// 获取piece begin idx 和 end idx
func (t *Torrent) calculateBoundsForPiece(index int) (int, int) {
	begin := index * t.PieceLength
	end := begin + t.PieceLength

	if end > t.Length {
		end = t.Length
	}

	return begin, end
}
