package bgp

import (
	"fmt"
	"net"
	"time"

	"github.com/osrg/gobgp/packet/bgp"
	"github.com/osrg/gobgp/table"
	v1 "k8s.io/api/core/v1"
)

type RouteInterface interface {
	Source() (net.IP, uint8)
	NextHop() net.IP
	Describe() string
	Path(bool) *table.Path
}

type Route struct {
	RouteInterface
}

func (r Route) String() string {
	prefix, length := r.Source()

	return fmt.Sprintf("%16s/%v -> %-15s (%s)", prefix.To4().String(), length, r.NextHop().To4().String(), r.Describe())
}

func (r Route) Path(isWithdraw bool) *table.Path {
	prefix, length := r.Source()
	nlri := bgp.NewIPAddrPrefix(length, prefix.To4().String())

	pattr := []bgp.PathAttributeInterface{
		bgp.NewPathAttributeOrigin(bgp.BGP_ORIGIN_ATTR_TYPE_IGP),
		bgp.NewPathAttributeNextHop(r.NextHop().To4().String()),
	}

	return table.NewPath(nil, nlri, isWithdraw, pattr, time.Now(), false)
}

type ExternalIPRoute struct {
	Route
	Service *v1.Service
	Proxy   *v1.Pod
}

func (r ExternalIPRoute) Source() (net.IP, uint8) {
	return net.ParseIP(r.Service.Spec.ExternalIPs[0]), uint8(32)
}

func (r ExternalIPRoute) NextHop() net.IP {
	return net.ParseIP(r.Proxy.Status.HostIP)
}

func (r ExternalIPRoute) Describe() string {
	return fmt.Sprintf("ExternalIP:    %s/%s -> %s/%s", r.Service.Namespace, r.Service.Name, r.Proxy.Namespace, r.Proxy.Name)
}

func NewExternalIPRoute(service *v1.Service, proxy *v1.Pod) RouteInterface {
	return ExternalIPRoute{Route{}, service, proxy}
}