package handshake

import (
	"fmt"
	"io"
)

// 握手报文
type HandShake struct {
	Pstr     string   // Bit Torrent Protocol
	InfoHash [20]byte // 验证hash
	PeerId   [20]byte // 客户端随机生成
}

func New(infohash, peerID [20]byte) *HandShake {
	return &HandShake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infohash,
		PeerId:   peerID,
	}
}

// 序列化
func (h *HandShake) Serialize() []byte {

	buf := make([]byte, len(h.Pstr)+49) //

	buf[0] = byte(len(h.Pstr))

	cur := 1

	cur += copy(buf[cur:], []byte(h.Pstr))
	cur += copy(buf[cur:], make([]byte, 8))
	cur += copy(buf[cur:], h.InfoHash[:])
	cur += copy(buf[cur:], h.PeerId[:])

	return buf
}

// 反序列化
func ReadHandShake(r io.Reader) (*HandShake, error) {
	// todo: 从与peers建立的连接中读取握手信息

	lengthBuf := make([]byte, 1)

	_, err := io.ReadFull(r, lengthBuf)

	if err != nil {
		return nil, err
	}

	// 消息长度
	pstrLen := int(lengthBuf[0])

	if pstrLen == 0 {
		err := fmt.Errorf("pstrLen can not be 0")
		return nil, err
	}

	handshakeBuf := make([]byte, 48+pstrLen)

	_, err = io.ReadFull(r, handshakeBuf)

	if err != nil {
		return nil, err
	}

	var infohash, peerID [20]byte

	// +8 忽略reserved信息
	copy(infohash[:], handshakeBuf[pstrLen+8:pstrLen+8+20])
	copy(peerID[:], handshakeBuf[pstrLen+20+8:])

	h := HandShake{
		Pstr:     string(handshakeBuf[0:pstrLen]),
		InfoHash: infohash,
		PeerId:   peerID,
	}

	return &h, nil
}
