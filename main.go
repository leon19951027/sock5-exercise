package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"time"
	"zjy-sock5/check"
	"zjy-sock5/tcpConnect"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var service string
	service = "0.0.0.0:8888"
	//返回 *TCPAddr 类型
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		panic(err)
	}
	//传入*TCPAddr类型
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}
	Checker := &check.CheckerImpl{}

	for {

		client, err := tcpListener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		clientConn := &tcpConnect.Client2SocksServerTcpConnectionImpl{
			TcpConnectionImpl: &tcpConnect.TcpConnectionImpl{Connection: client},
			Checker:           Checker,
		}

		buf := make([]byte, 1024)

		go func() {
			defer clientConn.Connection.Close()
			//第一个buf，客户端发送协议，认证方式等
			/*
						+----+----------+----------+
				   		|VER | NMETHODS |  METHODS |
				   		+----+----------+----------+
				   		|  1 |     1    | 1 to 255 |
				  		 +----+----------+----------+
			*/
			remoteAddrStr := clientConn.GetClientAddr()
			methodBufLength, methodBuf := clientConn.ReadBuf(buf)
			methodData := clientConn.GetBufData(methodBuf, methodBufLength)
			if !clientConn.Checker.CheckMethod(methodData) {
				fmt.Println(remoteAddrStr, methodData)
				log.Printf("WARN >>> from %s request methods invalid. \n ", remoteAddrStr)
				return
			} else {
				//校验通过，返回客户端buf数据
				clientConn.WriteBuf([]byte{0x05, 0x02})
			}

			//再次读取客户端给的buf，里面携带账号密码
			/*
				           VER	 IDLEN		 ID	       PWLEN			PW
				Byte count	1	  1	       (1-255)	    1			(1-255)
			*/
			authBufLength, authBuf := clientConn.ReadBuf(buf)
			authData := clientConn.GetBufData(authBuf, authBufLength)
			result, b0 := clientConn.Checker.CheckAuth(authData)
			if !result {
				log.Printf("WARN >>> from %s auth fail.\n", remoteAddrStr)
				return
			} else {
				clientConn.WriteBuf([]byte{b0, 0x00})
			}

			//校验工作完成，客户端发送目标地址
			// CMD 代表此连接是tcp还是udp，RSV固定，aypy代表ipv4或者ipv6或者域名
			/*
				+----+-----+-------+------+----------+--------+
				|VER | CMD | RSV | ATYP | DST.ADDR | DST.PORT |
				+----+-----+-------+------+----------+--------+
				|  1 |  1  |X’00’|  1   | Variable |     2    |
				+----+-----+-------+------+----------+--------+
			*/
			_, destnationAddrinfo := clientConn.ReadBuf(buf)
			addr, port := clientConn.GetAddrPort(destnationAddrinfo[3:])

			fmt.Println(addr, port)
			//拿到了目标地址，可以开始代理了
			switch buf[1] {
			case 0x01: //CONNECT -> TCP

				clientConn.WriteBuf([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

				remoteTcpConn, err := clientConn.DialRemote(addr, port)
				if err != nil {
					clientConn.WriteBuf([]byte{0x05, 0x03})
					return
				}

				socksServerConn := &tcpConnect.SocksServer2RemoteServerImpl{
					TcpConnectionImpl: &tcpConnect.TcpConnectionImpl{Connection: remoteTcpConn},
				}

				time.Sleep(10 * 1e6)

				clientConn.Connection.SetNoDelay(true)
				socksServerConn.Connection.SetNoDelay(true)

				closeSig := make(chan bool)
				go exchange(clientConn.Connection, socksServerConn.Connection, closeSig)
				go exchange(socksServerConn.Connection, clientConn.Connection, closeSig)
				fmt.Println(closeSig)
				<-closeSig
				return
			case 0x02:
				log.Println("WARN >>> get BIND command, not support.")
				return
			case 0x03: //UDP
				log.Println("WARN >>> udp proxy not support.")
				// p.udpProxy(dstAddr, dstPort)
				return
			default:
				return
			}
		}()
	}
}

func exchange(src, dst *net.TCPConn, closeSig chan bool) {
	fmt.Println("****************************")
	fmt.Println(src.RemoteAddr().String())
	fmt.Println(dst.RemoteAddr().String())
	buf := make([]byte, 0xff)
	for {
		n, err := src.Read(buf[0:])
		fmt.Println("//////////////////////////////////////////")
		fmt.Println(buf)
		if err != nil {
			fmt.Println(err)
			fmt.Println("退出")
			closeSig <- true
			return
		}
		b := buf[0:n]
		_, err = dst.Write(b)
		fmt.Println("\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\")
		fmt.Println(b)
		if err != nil {
			fmt.Println(err)
			fmt.Println("退出")
			closeSig <- true
			return
		}
	}
}
