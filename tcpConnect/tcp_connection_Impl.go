package tcpConnect

import (
	"fmt"
	"net"

	"zjy-sock5/check"
)

type TcpConnectionImpl struct {
	Connection *net.TCPConn
}

type Client2ServerTcpConnectionImpl struct {
	*TcpConnectionImpl
	Checker check.IChecker
}

type ITcpConnection interface {
	GetClientAddr() string
	ReadBuf(b []byte) (n int, buf []byte)
	GetBufData(b []byte, n int) (buf []byte)
	WriteBuf(b []byte)
	GetAddrPort(b []byte) (string, int)
}

func (t *TcpConnectionImpl) GetClientAddr() string {
	clientAddr := t.Connection.RemoteAddr().String()
	return clientAddr
}

func (t *TcpConnectionImpl) ReadBuf(b []byte) (n int, buf []byte) {
	n, err := t.Connection.Read(b[:])
	if err != nil {
		fmt.Println(err)
	}
	buf = b
	return n, buf
}

func (t *TcpConnectionImpl) GetBufData(b []byte, n int) (buf []byte) {
	data := b[0:n]
	return data
}

func (t *TcpConnectionImpl) WriteBuf(b []byte) {
	t.Connection.Write(b)
}

func (t *TcpConnectionImpl) GetAddrPort(b []byte) (string, int) {
	switch b[0] {
	case 0x01: //ipv4
		addr := net.IPv4(b[1], b[2], b[3], b[4]).String()
		port := int(b[5])*256 + int(b[6])
		return addr, port

	case 0x03: //domain
		domainLens := int(b[1])
		domain := string(b[2 : 2+domainLens])
		port := int(b[2+domainLens])*256 + int(b[2+domainLens+1])
		return domain, port
	default:
		return "", 0
	}
}
