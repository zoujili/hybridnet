/*
Copyright 2021 The Hybridnet Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"

	networkingv1 "github.com/alibaba/hybridnet/pkg/apis/networking/v1"
	"github.com/alibaba/hybridnet/pkg/daemon/iptables"
	"github.com/alibaba/hybridnet/pkg/daemon/neigh"
	"github.com/alibaba/hybridnet/pkg/daemon/route"
)

func (c *Controller) getRouterManager(ipVersion networkingv1.IPVersion) *route.Manager {
	if ipVersion == networkingv1.IPv6 {
		return c.routeV6Manager
	}
	return c.routeV4Manager
}

func (c *Controller) getNeighManager(ipVersion networkingv1.IPVersion) *neigh.Manager {
	if ipVersion == networkingv1.IPv6 {
		return c.neighV6Manager
	}
	return c.neighV4Manager
}

func (c *Controller) getIPtablesManager(ipVersion networkingv1.IPVersion) *iptables.Manager {
	if ipVersion == networkingv1.IPv6 {
		return c.iptablesV6Manager
	}
	return c.iptablesV4Manager
}

func (c *Controller) getIPInstanceByAddress(address net.IP) (*networkingv1.IPInstance, error) {
	ipInstanceList, err := c.ipInstanceIndexer.ByIndex(ByInstanceIPIndexer, address.String())
	if err != nil {
		return nil, fmt.Errorf("get ip instance by ip %v indexer failed: %v", address.String(), err)
	}

	if len(ipInstanceList) > 1 {
		return nil, fmt.Errorf("get more than one ip instance for ip %v", address.String())
	}

	if len(ipInstanceList) == 1 {
		instance, ok := ipInstanceList[0].(*networkingv1.IPInstance)
		if !ok {
			return nil, fmt.Errorf("transform obj to ipinstance failed")
		}

		return instance, nil
	}

	if len(ipInstanceList) == 0 {
		// not found
		return nil, nil
	}

	return nil, fmt.Errorf("ip instance for address %v not found", address.String())
}

func initErrorMessageWrapper(prefix string) func(string, ...interface{}) string {
	return func(format string, args ...interface{}) string {
		return prefix + fmt.Sprintf(format, args...)
	}
}

func parseSubnetSpecRangeMeta(addressRange *networkingv1.AddressRange) (cidr *net.IPNet, gateway, start, end net.IP,
	excludeIPs, reservedIPs []net.IP, err error) {

	if addressRange == nil {
		return nil, nil, nil, nil, nil, nil,
			fmt.Errorf("cannot parse a nil range")
	}

	cidr, err = netlink.ParseIPNet(addressRange.CIDR)
	if err != nil {
		return nil, nil, nil, nil, nil, nil,
			fmt.Errorf("failed to parse subnet cidr %v error: %v", addressRange.CIDR, err)
	}

	gateway = net.ParseIP(addressRange.Gateway)
	if gateway == nil {
		return nil, nil, nil, nil, nil, nil,
			fmt.Errorf("invalid gateway ip %v", addressRange.Gateway)
	}

	if addressRange.Start != "" {
		start = net.ParseIP(addressRange.Start)
		if start == nil {
			return nil, nil, nil, nil, nil, nil,
				fmt.Errorf("invalid start ip %v", addressRange.Start)
		}
	}

	if addressRange.End != "" {
		end = net.ParseIP(addressRange.End)
		if end == nil {
			return nil, nil, nil, nil, nil, nil,
				fmt.Errorf("invalid end ip %v", addressRange.End)
		}
	}

	for _, ipString := range addressRange.ExcludeIPs {
		excludeIP := net.ParseIP(ipString)
		if excludeIP == nil {
			return nil, nil, nil, nil, nil, nil,
				fmt.Errorf("invalid exclude ip %v", ipString)
		}
		excludeIPs = append(excludeIPs, excludeIP)
	}

	for _, ipString := range addressRange.ReservedIPs {
		reservedIP := net.ParseIP(ipString)
		if reservedIP == nil {
			return nil, nil, nil, nil, nil, nil,
				fmt.Errorf("invalid reserved ip %v", ipString)
		}
		reservedIPs = append(reservedIPs, reservedIP)
	}

	return
}
