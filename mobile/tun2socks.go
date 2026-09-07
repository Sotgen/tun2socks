package mobile

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/fdbased"
	_ "github.com/xjasonlyu/tun2socks/v2/proxy/socks"
)

var (
	mu      sync.Mutex
	running bool
	dev     device.Device
)

// StartTun2Socks menjalankan tunnel gVisor menggunakan FD dari Android VpnService
// fd: file descriptor integer dari Android
// socksAddr: contoh "127.0.0.1:1080"
// mtu: standar 1500
func StartTun2Socks(fd int, socksAddr string, mtu int) error {
	mu.Lock()
	defer mu.Unlock()

	if running {
		return errors.New("tun2socks is already running")
	}

	// 1. Parameter fdbased.Open pada tun2socks v2: (name string, mtu uint32, fd int)
	tunDev, err := fdbased.Open("tun", uint32(mtu), fd)
	if err != nil {
		return fmt.Errorf("failed to open fd: %w", err)
	}
	dev = tunDev

	// 2. Format URL proxy SOCKS5
	proxyURL := fmt.Sprintf("socks5://%s", socksAddr)

	// 3. Register handler dan device ke core engine
	core.RegisterOutputDevice(dev)

	// 4. Start core engine tun2socks (gVisor stack)
	if err := core.Start(core.WithProxy(proxyURL), core.WithStack("gvisor")); err != nil {
		dev.Close()
		dev = nil
		return fmt.Errorf("failed to start core: %w", err)
	}

	running = true
	return nil
}

// StopTun2Socks mematikan service VPN
func StopTun2Socks() {
	mu.Lock()
	defer mu.Unlock()

	if !running {
		return
	}

	core.Stop()

	if dev != nil {
		dev.Close()
		dev = nil
	}

	running = false
}

// IsRunning cek status service
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return running
}
