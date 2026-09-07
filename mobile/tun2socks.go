package mobile

import (
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/fdbased"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
)

var (
	mu      sync.Mutex
	running bool
	dev     device.Device
)

// StartTun2Socks menerima File Descriptor dari VpnService Android dan proxy lokal
// fd: file descriptor integer dari Android
// socksAddr: format "127.0.0.1:1080"
// mtu: biasanya 1500
func StartTun2Socks(fd int, socksAddr string, mtu int) error {
	mu.Lock()
	defer mu.Unlock()

	if running {
		return errors.New("tun2socks is already running")
	}

	// 1. Buka device TUN dari FD Android
	tunDev, err := fdbased.Open(uintptr(fd), uint32(mtu))
	if err != nil {
		return fmt.Errorf("failed to open fd: %w", err)
	}
	dev = tunDev

	// 2. Parse URL proxy SOCKS5
	proxyURI, err := url.Parse("socks5://" + socksAddr)
	if err != nil {
		_ = dev.Close()
		return fmt.Errorf("invalid socks address: %w", err)
	}

	handler, err := proxy.NewProxy(proxyURI)
	if err != nil {
		_ = dev.Close()
		return fmt.Errorf("failed to create proxy handler: %w", err)
	}

	// 3. Jalankan core engine dengan stack gvisor
	err = core.Start(
		core.WithDevice(dev),
		core.WithProxyHandler(handler),
		core.WithStack("gvisor"),
	)
	if err != nil {
		_ = dev.Close()
		return fmt.Errorf("failed to start core: %w", err)
	}

	running = true
	return nil
}

// StopTun2Socks untuk mematikan engine
func StopTun2Socks() {
	mu.Lock()
	defer mu.Unlock()

	if !running {
		return
	}

	if dev != nil {
		_ = dev.Close()
		dev = nil
	}
	core.Stop()
	running = false
}

// IsRunning cek status koneksi
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return running
}
