package conn

import (
	"encoding/binary"
	"io"
)

type messageID uint8

// typeid
const (
	MsgChoke         messageID = 0
	MsgUnchoke       messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

// 实际传输的数据
type Message struct {
	ID      messageID
	PayLoad []byte
}

// 序列化msg
func (m *Message) Serialize() []byte {
	// todo: 序列化msg传输
	if m == nil {
		return make([]byte, 4)
	}

	length := uint32(len(m.PayLoad) + 1)

	buf := make([]byte, 4+length)

	binary.BigEndian.PutUint32(buf[0:4], length)

	// msg typid
	buf[4] = byte(m.ID)

	// msg payload
	copy(buf[5:], m.PayLoad)

	return buf

}

// 反序列化msg
func ReadMessage(r io.Reader) (*Message, error) {
	// todo: 通过msg offset从stream中反序列化msg

	lengthBuf := make([]byte, 4) // 获取msg length prefix

	_, err := io.ReadFull(r, lengthBuf)

	if err != nil {
		return nil, err
	}

	// 获取msg长度
	length := binary.BigEndian.Uint32(lengthBuf)

	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)

	_, err = io.ReadFull(r, messageBuf)

	if err != nil {
		return nil, err
	}

	m := &Message{
		ID:      messageID(messageBuf[0]),
		PayLoad: messageBuf[1:],
	}

	return m, nil
}

// bitmap
type BitField []byte

// 判断是否存在piece
func (bf BitField) HasPiece(idx int) bool {

	// peice位于整个数组中的下标
	byteIdx := idx / 8

	// piece在对应下标的位置
	offset := idx % 8

	// 每个字节分别存储8个piece
	// 对应字节位位1代表存在piece
	// 右移7-offset正好可以得到对应piece的位
	// 1 & 1 = 1  0 & 1 = 0
	return bf[byteIdx]>>(7-offset)&1 != 0
}

// 设置对应字节位上的piece存在
func (bf BitField) SetPiece(idx int) {

	byteIdx := idx / 8

	offset := idx % 8

	// 相反的设置对应piece的位
	// 则是左移相或
	// 1 | 1 = 1  1 | 0 = 1
	bf[byteIdx] |= 1 << (7 - offset)
}
