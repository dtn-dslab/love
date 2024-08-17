package device

import (
	"fmt"
	"net"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
	qdisc "github.com/dtn-dslab/love/pkg/qdisc"
	"github.com/vishvananda/netlink"
)

type Device struct {
	// Name of the network device
	Name string
	// Address of the network device
	Address string
	// IP address of the network device
	IPAddr []net.IPNet
	// MTU of the network device
	MTU int
	// Network namespace of the network device
	Namespace string
	// Properties of the network device
	Properties *qdisc.DeviceProperties
	// netlink interface
	Link netlink.Link
}

func (d *Device) GetNs() (ns.NetNS, error) {
	if d.Namespace == "" {
		return ns.GetCurrentNS()
	}
	return ns.GetNS(d.Namespace)
}

func (d *Device) OpenDevice() (err error) {
	var vethNs ns.NetNS
	if vethNs, err = d.GetNs(); err != nil {
		return err
	}
	defer vethNs.Close()

	err = vethNs.Do(func(_ ns.NetNS) error {
		link, err := netlink.LinkByName(d.Name)
		if err != nil {
			return err
		}
		d.Link = link

		return nil
	})

	return err
}

func (d *Device) DeleteDevice() (err error) {
	var vethNs ns.NetNS
	if vethNs, err = d.GetNs(); err != nil {
		return err
	}
	defer vethNs.Close()

	err = vethNs.Do(func(_ ns.NetNS) error {
		return netlink.LinkDel(d.Link)
	})

	return err
}

// SetDevice sets the device with
func (d *Device) SetDevice() (err error) {
	var vethNs ns.NetNS
	if vethNs, err = d.GetNs(); err != nil {
		return err
	}
	defer vethNs.Close()

	err = vethNs.Do(func(_ ns.NetNS) error {
		if err := netlink.LinkSetUp(d.Link); err != nil {
			return err
		}

		if err := netlink.LinkSetMTU(d.Link, d.MTU); err != nil {
			return err
		}

		hardwareAddr, _ := net.ParseMAC(d.Address)
		if err := netlink.LinkSetHardwareAddr(d.Link, hardwareAddr); err != nil {
			return err
		}

		// Clean all IP addresses, except for IPv6 link-local
		addrs, err := netlink.AddrList(d.Link, netlink.FAMILY_ALL)
		if err != nil {
			return fmt.Errorf("failed to list addresses on %q: %v",
				d.Name, err)
		}
		for _, addr := range addrs {
			if addr.IP.IsLinkLocalUnicast() {
				continue
			}
			if err = netlink.AddrDel(d.Link, &addr); err != nil {
				return fmt.Errorf("failed to remove address %v from %q: %v",
					addr, d.Name, err)
			}
		}

		// Conditionally set the IP address.
		for i := 0; i < len(d.IPAddr); i++ {
			// if IPv6, need to enable IPv6 using sysctl
			if d.IPAddr[i].IP.To4() == nil {
				ipv6SysctlName := fmt.Sprintf("net.ipv6.conf.%s.disable_ipv6",
					d.Name)
				if _, err := sysctl.Sysctl(ipv6SysctlName, "0"); err != nil {
					return fmt.Errorf("failed to set ipv6.disable to 0 at %s: %v",
						d.Name, err)
				}

			}
			addr := &netlink.Addr{IPNet: &d.IPAddr[i], Label: ""}
			if err = netlink.AddrAdd(d.Link, addr); err != nil {
				return fmt.Errorf(
					"failed to add IP addr %v to %q: %v",
					addr, d.Name, err)
			}
		}

		return nil
	})

	return err

}

func (d *Device) SetDeviceProperties() (err error) {
	qdiscs, err := d.Properties.ParseQdiscs()
	if err != nil {
		return err
	}

	if err := d.ClearDeviceProperties(); err != nil {
		return err
	}

	var vethNs ns.NetNS
	if vethNs, err = d.GetNs(); err != nil {
		return err
	}
	defer vethNs.Close()

	err = vethNs.Do(func(_ ns.NetNS) error {
		withNetem := false
		if len(qdiscs) == 2 {
			withNetem = true
		}

		for _, qdisc := range qdiscs {
			// Set link index and parent, assign handle
			var newQdisc netlink.Qdisc
			switch qdisc := qdisc.(type) {
			case *netlink.Netem:
				qdisc.LinkIndex = d.Link.Attrs().Index
				qdisc.Parent = netlink.HANDLE_ROOT
				qdisc.Handle = netlink.MakeHandle(1, 0) // 1:0
				newQdisc = qdisc
				err = netlink.QdiscAdd(newQdisc)
			case *netlink.Tbf:
				qdisc.LinkIndex = d.Link.Attrs().Index
				if withNetem {
					qdisc.Parent = netlink.MakeHandle(1, 1)  // 1:1
					qdisc.Handle = netlink.MakeHandle(10, 0) // 10:0
				} else {
					qdisc.Parent = netlink.HANDLE_ROOT
				}
				newQdisc = qdisc
				err = netlink.QdiscAdd(newQdisc)
			default:
				fmt.Errorf("Unsupported qdisc type %s", qdisc.Type())
			}

			if err != nil {
				fmt.Errorf("Failed to set qdisc %v to link %s: %v", qdisc, d.Name, err)
				return err
			}
		}

		return nil
	})

	return err
}

func (d *Device) ClearDeviceProperties() (err error) {
	var vethNs ns.NetNS
	if vethNs, err = d.GetNs(); err != nil {
		return err
	}
	defer vethNs.Close()

	err = vethNs.Do(func(_ ns.NetNS) error {
		qdiscs, err := netlink.QdiscList(d.Link)
		if err != nil {
			fmt.Errorf("Failed to list qdiscs for link %s: %v", d.Name, err)
			return err
		}

		for _, qdisc := range qdiscs {
			if err := netlink.QdiscDel(qdisc); err != nil {
				fmt.Errorf("Failed to delete qdisc %s: %v", qdisc.Type(), err)
				return err
			}
		}

		return nil
	})

	return err
}
