package ipc

import (
	"fmt"
	"net"
)

func ConnectToPipe(ipcSocketPath string) (net.Conn, error) {
	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		return nil, fmt.Errorf("ipc bağlantısı kurulamadı: %w", err)
	}
	return conn, nil
}
