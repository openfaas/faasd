// +build linux

package cninetwork

import "github.com/vishvananda/netlink"

func linkToNetDev(link netlink.Link) (Dev, error) {

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return Dev{}, err
	}

	netDev := Dev{Name: link.Attrs().Name, MAC: link.Attrs().HardwareAddr}
	for _, addr := range addrs {
		netDev.CIDRs = append(netDev.CIDRs, addr.IPNet)
	}
	return netDev, nil
}
