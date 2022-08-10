package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

// typeid
const (
	MsgChoke         messageID = 0 //type id = 0 block msg
	MsgUnchoke       messageID = 1 //type id = 1 unblock msg
	MsgInterested    messageID = 2 //type id = 2
	MsgNotInterested messageID = 3 //type id = 3
	MsgHave          messageID = 4 //type id = 4 表示当前终端下载了对应的piece payload中包含该piece的sha1校验值
	MsgBitfield      messageID = 5 //type id = 5 bitfied msg payload的数据是包含piece信息的 bitfield
	MsgRequest       messageID = 6 //type id = 6 request msg payload的数据是index, begin, length 分别代表文件分片的索引，对应piece内的字节索引, 请求的长度
	MsgPiece         messageID = 7 //type id = 7 piece msg payload的数据是index, begin, piece 前两个的意义与request msg相同， piece则是对端peer请求的文件片段
	MsgCancel        messageID = 8 //type id = 8 cancel msg payload数据是index, begin, length 意义与request msg 相反 用于取消对应文件片段的下载
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

// 创建 MsgReq
func FormatRequest(idx, begin, length int) (*Message, error) {
	payLoad := make([]byte, 12)

	binary.BigEndian.PutUint32(payLoad[:4], uint32(idx))
	binary.BigEndian.PutUint32(payLoad[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payLoad[8:12], uint32(length))

	return &Message{
		ID:      MsgRequest,
		PayLoad: payLoad,
	}, nil
}

// 创建 MsgHave
func FromHava(idx int) *Message {
	payload := make([]byte, 4)

	binary.BigEndian.PutUint32(payload, uint32(idx))

	return &Message{ID: MsgHave, PayLoad: payload}
}

// parse 对等peer发送的Msgpiece
//
// 返回data长度
func ParseMsgPiece(idx int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		err := fmt.Errorf("expect Piece Msg, ID:%d but got:%d", MsgPiece, msg.ID)
		return 0, err
	}

	if len(msg.PayLoad) < 8 {
		err := fmt.Errorf("msg payload is to short, length:%d", len(msg.PayLoad))
		return 0, err
	}

	// 校验请求的idx是否与返回的相同
	pieceIdx := int(binary.BigEndian.Uint32(msg.PayLoad[0:4]))

	if pieceIdx != idx {
		err := fmt.Errorf("expect piece idx is :%d, but got:%d", idx, pieceIdx)
		return 0, err
	}

	begin := int(binary.BigEndian.Uint32(msg.PayLoad[4:8]))

	if begin >= len(buf) {
		err := fmt.Errorf("begin offset is too big, begin:%d, bufLength:%d", begin, len(buf))
		return 0, err
	}

	data := msg.PayLoad[8:]

	if begin+len(data) >= len(buf) {
		err := fmt.Errorf("data is too long for buf, datasize:%d, bufsize:%d, beiginoffset:%d", len(data), len(buf), begin)
		return 0, err
	}

	copy(buf[begin:], data)

	return len(data), nil
}

// parse 对等peer发送的MsgHave
func ParseMsgHave(msg *Message) (int, error) {
	if msg.ID != MsgHave {
		err := fmt.Errorf("expect MsgHave, ID:%d but got:%d", MsgHave, msg.ID)
		return 0, err
	}

	if len(msg.PayLoad) != 4 {
		err := fmt.Errorf("expected payload length is 4 but got:%d", len(msg.PayLoad))
		return 0, err
	}

	idx := int(binary.BigEndian.Uint32(msg.PayLoad))

	return idx, nil
}

func (m *Message) name() string {
	if m == nil {
		return "keep alive"
	}

	switch m.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "NotInterested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown#%d", m.ID)
	}
}

func (m *Message) String() string {

	if m == nil {
		return m.name()
	}

	fmtstr := "%s [%d]"

	return fmt.Sprintf(fmtstr, m.name(), len(m.PayLoad))

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
