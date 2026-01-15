package nmlite

import "github.com/xkvm/kvm/pkg/nmlite/link"

func getNetlinkManager() *link.NetlinkManager {
	return link.GetNetlinkManager()
}
