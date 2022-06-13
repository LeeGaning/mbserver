package mbserver

import (
	"bytes"
	"io"
	"log"

	"github.com/goburrow/serial"
)

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTU(serialConfig *serial.Config) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Fatalf("failed to open %s: %v\n", serialConfig.Address, err)
	}
	s.ports = append(s.ports, port)
	go s.acceptSerialRequests(port)
	return err
}

func (s *Server) acceptSerialRequests(port serial.Port) {
	buffer := bytes.Buffer{}
SkipFrameError:
	for {
		buf := make([]byte, 512)
		bytesRead, err := port.Read(buf)
		if err != nil {
			if err != io.EOF {
				for buffer.Len() > 5 {
					b := make([]byte, buffer.Len())
					_, err := buffer.Read(b)
					if err != nil {
						log.Printf("buffer read error %v\n", err)
						break
					}
					frame, err := NewRTUFrame(b)
					if err != nil {
						log.Printf("bad serial frame error %v\n", err)
						// return
						break
					}
					request := &Request{port, frame}
					// s.requestChan <- request
					response := s.handle(request)
					port.Write(response.Bytes())
				}
				continue SkipFrameError
			}
			log.Fatalf("serial read error %v\n", err)
			return
		}
		buffer.Write(buf[:bytesRead])
	}
}
