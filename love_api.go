package love

import (
	"fmt"
	"net"
	"os"

	device "github.com/dtn-dslab/love/pkg/device"
	qdisc "github.com/dtn-dslab/love/pkg/qdisc"
	"github.com/vishvananda/netlink"
)

const (
	// DefaultMTU is the default MTU for a network device
	DefaultMTU = 1500
)

type VEth struct {
	// Local network device
	LocalDevice device.Device
	// Peer network device
	PeerDevice device.Device
}

func MakeVEth(config VEthConfig) *VEth {
	ve := &VEth{
		LocalDevice: device.Device{
			Name:       config.LocalName,
			Address:    config.LocalAddr,
			IPAddr:     config.LocalIPAddr,
			Namespace:  config.LocalNamspace,
			Properties: config.LocalProps,
		},
		PeerDevice: device.Device{
			Name:       config.PeerName,
			Address:    config.PeerAddr,
			IPAddr:     config.PeerIPAddr,
			Namespace:  config.PeerNamespace,
			Properties: config.PeerProps,
		},
	}

	if config.MTU == 0 {
		ve.LocalDevice.MTU = DefaultMTU
		ve.PeerDevice.MTU = DefaultMTU
	} else {
		ve.LocalDevice.MTU = config.MTU
		ve.PeerDevice.MTU = config.MTU
	}

	return ve
}

func (v *VEth) AddVEth() error {
	localHardwareAddr, _ := net.ParseMAC(v.LocalDevice.Address)
	peerHardwareAddr, _ := net.ParseMAC(v.PeerDevice.Address)

	localLink := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			MTU:          v.LocalDevice.MTU,
			Name:         v.LocalDevice.Name,
			HardwareAddr: localHardwareAddr,
		},
		PeerName:         v.PeerDevice.Name,
		PeerHardwareAddr: peerHardwareAddr,
	}

	if err := netlink.LinkAdd(localLink); err != nil {
		switch {
		case os.IsExist(err):
			// TODO: Define a custom error type
			return fmt.Errorf(
				"container veth name provided (%v) "+
					"already exists", v.LocalDevice.Name)
		default:
			return fmt.Errorf("failed to make veth pair: %v", err)

		}
	}
	v.LocalDevice.Link = localLink

	if err := v.LocalDevice.OpenDevice(); err != nil {
		return fmt.Errorf("failed to open local device: %v", err)
	}

	return nil
}

func (v *VEth) DeleteVEth() error {
	if err := v.LocalDevice.DeleteDevice(); err != nil {
		return fmt.Errorf("failed to delete local device: %v", err)
	}

	if err := v.PeerDevice.DeleteDevice(); err != nil {
		return fmt.Errorf("failed to delete peer device: %v", err)
	}

	return nil
}

func (v *VEth) OpenVEth() error {
	if err := v.LocalDevice.OpenDevice(); err != nil {
		return fmt.Errorf("failed to open local device: %v", err)
	}

	if err := v.PeerDevice.OpenDevice(); err != nil {
		return fmt.Errorf("failed to open peer device: %v", err)
	}

	return nil
}

type VEthConfig struct {
	MTU           int
	LocalName     string
	LocalAddr     string
	LocalIPAddr   []net.IPNet
	LocalNamspace string
	PeerName      string
	PeerAddr      string
	PeerIPAddr    []net.IPNet
	PeerNamespace string
	LocalProps    *qdisc.DeviceProperties
	PeerProps     *qdisc.DeviceProperties
}
