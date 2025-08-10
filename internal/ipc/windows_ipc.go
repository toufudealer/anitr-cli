//go:build windows
// +build windows

package ipc

import (
	"net"

	"github.com/Microsoft/go-winio"
)

// ConnectToPipe, verilen NPIPE soket yoluna bağlanmaya çalışır.
// Başarılı olursa net.Conn nesnesi döner, aksi hâlde hata döner.
func ConnectToPipe(ipcSocketPath string) (net.Conn, error) {
	conn, err := winio.DialPipe(ipcSocketPath, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
