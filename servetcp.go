package mbserver

import (
	"io"
	"log"
	"net"
	"strings"
)

func (s *Server) accept(listen net.Listener) error {
	for {
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			log.Printf("Unable to accept connections: %#v\n", err)
			return err
		}

		go func(conn net.Conn) {
			log.Printf("connected: %v<->%v\n", conn.LocalAddr(), conn.RemoteAddr())
			packet_count := 0
			packet := make([]byte, 512)
			defer func() {
				conn.Close()
				log.Printf("closed: %v<->%v,packet_count:%d\n", conn.LocalAddr(), conn.RemoteAddr(), packet_count)
			}()
			for {
				bytesRead, err := conn.Read(packet)
				if err != nil {
					if err != io.EOF {
						log.Printf("read error %v<->%v, %v\n", conn.LocalAddr(), conn.RemoteAddr(), err)
					}
					return
				}
				// log.Printf("packet %v,%d [% x]\n", conn.RemoteAddr(), bytesRead, packet[:bytesRead])
				// Set the length of the packet to the number of read bytes.
				frame, err := NewTCPFrame(packet[:bytesRead])
				if err != nil {
					log.Printf("bad packet error %v\n", err)
					return
				}
				packet_count++
				request := &Request{conn, frame}

				// s.requestChan <- request
				response := s.handle(request)
				conn.Write(response.Bytes())
			}
		}(conn)
	}
}

// ListenTCP starts the Modbus server listening on "address:port".
func (s *Server) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Printf("Failed to Listen: %v\n", err)
		return err
	}
	s.listeners = append(s.listeners, listen)
	go s.accept(listen)
	return err
}
