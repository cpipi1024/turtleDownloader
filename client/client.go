package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"cpipi1024.com/turtleDownloader/utils/bitfield"
	"cpipi1024.com/turtleDownloader/utils/handshake"
	"cpipi1024.com/turtleDownloader/utils/message"
	"cpipi1024.com/turtleDownloader/utils/peers"
)

// peer to peer TCP通信客户端
type Client struct {
	Conn     net.Conn          // tcp connection对象
	Choked   bool              // 通信阻塞标志
	BitField bitfield.BitField // peer承载数据的bitmap
	peer     peers.Peer
	infohash [20]byte
	peerId   [20]byte
}

// peer进行握手
func completeHandShake(conn net.Conn, infohash, peerID [20]byte) (*handshake.HandShake, error) {

	conn.SetDeadline(time.Now().Add(15 * time.Second))

	defer conn.SetDeadline(time.Time{})

	req := handshake.New(infohash, peerID)

	// 发送握手消息
	_, err := conn.Write(req.Serialize())

	if err != nil {
		return nil, err
	}

	// peer响应信息
	hsReponse, err := handshake.ReadHandShake(conn)

	if err != nil {
		return nil, err
	}

	if !bytes.Equal(hsReponse.InfoHash[:], infohash[:]) {
		err := fmt.Errorf("file infohash is not match want:%v ,got:%v", infohash, hsReponse.InfoHash)
		return nil, err
	}

	return hsReponse, nil
}

// 从连接中读取MsgBitField
func reciveBitField(conn net.Conn) (bitfield.BitField, error) {
	conn.SetDeadline(time.Now().Add(15 * time.Second))

	defer conn.SetDeadline(time.Time{})

	msg, err := message.ReadMessage(conn)

	if err != nil {
		return nil, err
	}

	if msg == nil {
		err := fmt.Errorf("client read bitfield msg failed")
		return nil, err
	}

	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("expect bitfield msg ID:%d but got:%d", message.MsgBitfield, msg.ID)
		return nil, err
	}

	return msg.PayLoad, nil
}

// 批量创建多个client与传入的peer通信
func NewClient(peer peers.Peer, peerID, infohash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 5*time.Second)
	if err != nil {
		return nil, err
	}

	// 先与peer进行握手
	_, err = completeHandShake(conn, infohash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// 接受 msgBitFiled
	bf, err := reciveBitField(conn)

	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   false,
		BitField: bf,
		peer:     peer,
		infohash: infohash,
		peerId:   peerID,
	}, nil
}

// 客户端读取的消息
func (c *Client) Read() (*message.Message, error) {
	msg, err := message.ReadMessage(c.Conn)
	return msg, err
}

// 客户端发送请求消息
func (c *Client) SendRequest(begin, idx, length int) error {
	req := message.FormatRequest(begin, idx, length)

	_, err := c.Conn.Write(req.Serialize())

	return err
}

func (c *Client) SendInterested() error {
	m := message.Message{ID: message.MsgInterested}

	_, err := c.Conn.Write(m.Serialize())

	return err
}

func (c *Client) SendNotInterested() error {
	m := message.Message{ID: message.MsgNotInterested}

	_, err := c.Conn.Write(m.Serialize())

	return err
}
func (c *Client) SendUnchoke() error {
	m := message.Message{ID: message.MsgUnchoke}

	_, err := c.Conn.Write(m.Serialize())

	return err
}

func (c *Client) SendHave(idx int) error {
	m := message.FromHava(idx)

	_, err := c.Conn.Write(m.Serialize())

	return err
}
