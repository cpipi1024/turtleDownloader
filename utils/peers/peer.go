package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Peer struct {
	IP   net.IP // peer ip地址
	Port uint   // peer 端口
}

func (p Peer) String() string {
	// 返回id:port字符串
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

func Unmarshal(peersData []byte) ([]Peer, error) {

	//todo: 从bt traker的响应报文获取peers
	const peerSize = 6

	peerLength := len(peersData)

	if peerLength%peerSize != 0 {
		err := fmt.Errorf("reciver malformed peers")

		return nil, err
	}

	peerNum := peerLength / peerSize

	peers := make([]Peer, peerNum)

	for i := 0; i < peerNum; i++ {
		offset := i * peerSize

		peers[i].IP = net.IP(peersData[offset : offset+4])
		peers[i].Port = uint(binary.BigEndian.Uint16(peersData[offset+4 : offset+6]))
	}

	return peers, nil

}
