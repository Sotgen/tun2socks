package mobile

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xjasonlyu/tun2socks/v2/engine"
)

var (
	mu      sync.Mutex
	running bool
	key     *engine.Key
)

// StartTun2Socks menjalankan engine tun2socks menggunakan FD dari Android VpnService
func StartTun2Socks(fd int, socksAddr string, mtu int) error {
	mu.Lock()
	defer mu.Unlock()

	if running {
		return errors.New("tun2socks is already running")
	}

	// Device string format URL fd bawaan engine: fd://<angka>
	k := &engine.Key{
		Device: fmt.Sprintf("fd://%d", fd),
		Proxy:  fmt.Sprintf("socks5://%s", socksAddr),
		MTU:    mtu,
	}

	engine.Insert(k)
	engine.Start()

	key = k
	running = true
	return nil
}

// StopTun2Socks mematikan tunnel
func StopTun2Socks() {
	mu.Lock()
	defer mu.Unlock()

	if !running {
		return
	}

	engine.Stop()
	key = nil
	running = false
}

// IsRunning cek status service
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return running
}
