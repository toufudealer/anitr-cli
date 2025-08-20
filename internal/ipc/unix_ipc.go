//go:build !windows
// +build !windows

package ipc

import (
	"fmt"
	"net"
)

// ConnectToPipe, verilen UNIX soket yoluna bağlanmaya çalışır.
// Başarılı olursa net.Conn nesnesi döner, aksi hâlde hata döner.
func ConnectToPipe(ipcSocketPath string) (net.Conn, error) {
	// UNIX soketi üzerinden bağlantı kurmayı dene
	conn, err := net.Dial("unix", ipcSocketPath)
	if err != nil {
		// Bağlantı kurulamazsa hata mesajıyla birlikte döndür
		return nil, fmt.Errorf("ipc bağlantısı kurulamadı: %w", err)
	}

	// Bağlantı başarılıysa geri döndür
	return conn, nil
}
