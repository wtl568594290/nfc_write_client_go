package views

import (
	"errors"
	"time"

	"go.bug.st/serial"
)

var Port serial.Port

func PortWriteAndRead(data string, ms int) (string, error) {
	if Port == nil {
		return "", errors.New("port is nil")
	}
	Port.ResetInputBuffer()
	Port.ResetOutputBuffer()
	if _, err := Port.Write([]byte(data)); err != nil {
		return "", err
	}
	buff := make([]byte, 1024)
	done := make(chan struct{})
	_portCache := ""
	go func() {
		defer close(done)
		for {
			n, err := Port.Read(buff)
			if err != nil {
				return
			}
			if n == 0 {
				continue
			}
			for i := 0; i < n; i++ {
				if buff[i] == '\n' {
					_portCache += string(buff[:i])
					return
				}
			}
			_portCache += string(buff[:n])
		}
	}()
	select {
	case <-done:
		return _portCache, nil
	case <-time.After(time.Duration(ms) * time.Millisecond):
		return "", errors.New("read timeout")
	}
}
