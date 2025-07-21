package ipc

import "net"

func ConnectToPipe(ipcSocketPath string) (net.Conn, error) {
	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
