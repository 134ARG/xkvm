package nmlite

import "github.com/134ARG/xkvm/pkg/nmlite/link"

func getNetlinkManager() *link.NetlinkManager {
	return link.GetNetlinkManager()
}
