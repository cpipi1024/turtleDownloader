package conn

import "io"

// 握手报文
type HandShake struct {
	Pstr     string   // Bit Torrent Protocol
	InfoHash [20]byte // 验证hash
	PeerId   [20]byte // 客户端随机生成
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
	// todo: 从与tracker建立的连接中读取握手信息

	return nil, nil
}
