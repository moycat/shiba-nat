package main

import (
	"fmt"
	"net"
	"syscall"

	"github.com/moycat/shiba-nat/internal"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func findRoute(ip net.IP) (netlink.Link, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}
	for _, route := range routes {
		if route.Dst == nil {
			continue
		}
		if !route.Dst.Contains(ip) {
			continue
		}
		link, err := netlink.LinkByIndex(route.LinkIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to get link by index [%d]: %w", route.LinkIndex, err)
		}
		log.Debugf("found route to [%s] on link [%s]", ip.String(), link.Attrs().Name)
		return link, nil
	}
	return nil, fmt.Errorf("failed to find a route to [%s]: %w", ip.String(), err)
}

func setDefaultRoute(link netlink.Link) error {
	linkName, linkIndex := link.Attrs().Name, link.Attrs().Index
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return fmt.Errorf("failed to list routes: %w", err)
	}
	// Check all existing routes, and remove unwanted ones.
	for _, route := range routes {
		if !isDefaultRoute(&route) {
			continue
		}
		if route.LinkIndex == linkIndex {
			// Already set.
			log.Debugf("default route via [%s] already set", linkName)
			return nil
		}
		if err := netlink.RouteDel(&route); err != nil {
			return fmt.Errorf("failed to delete unwanted route [%s]: %w", route.String(), err)
		}
		log.Debugf("deleted unwanted route [%s]", route.String())
	}
	// Add the default route via the given device.
	route := &netlink.Route{
		LinkIndex: linkIndex,
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
		Scope:     syscall.RT_SCOPE_LINK,
	}
	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("failed to add the default route [%s]: %w", route.String(), err)
	}
	log.Infof("set default route via [%s]", linkName)
	return nil
}

func isDefaultRoute(route *netlink.Route) bool {
	if route.Dst == nil {
		route.Dst = &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)} // Add a placeholder.
		return true
	}
	if !route.Dst.IP.Equal(net.IPv4zero) {
		return false
	}
	if ones, bits := route.Dst.Mask.Size(); ones != 0 || bits != 32 {
		return false
	}
	return true
}

func generateQuery(port int) *internal.Query {
	return &internal.Query{
		Magic: internal.QueryMagic,
		Token: xid.New().String(),
		Port:  port,
	}
}
