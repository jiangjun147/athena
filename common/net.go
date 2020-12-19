package common

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func GetLocalAddr() string {
	addrs, err := net.InterfaceAddrs()
	AssertError(err)

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func GetHardwareAddr() net.HardwareAddr {
	ifaces, err := net.Interfaces()
	AssertError(err)

	for _, iface := range ifaces {
		if (iface.Flags&net.FlagUp) != 0 && (iface.Flags&net.FlagLoopback) == 0 {
			addrs, _ := iface.Addrs()
			for _, address := range addrs {
				ipnet, ok := address.(*net.IPNet)
				if ok && ipnet.IP.IsGlobalUnicast() {
					return iface.HardwareAddr
				}
			}
		}
	}
	return nil
}

func AddressToIp4Port(address string) (ip4 string, port int, err error) {
	ss := strings.Split(address, ":")
	if len(ss) != 2 {
		err = fmt.Errorf("address invalid: %s", address)
		return
	}
	ip4 = ss[0]

	n, err := strconv.ParseInt(ss[1], 10, 64)
	if err != nil {
		return
	}

	port = int(n)
	return
}

func Ip4PortToAddress(ip4 string, port int) (address string) {
	return fmt.Sprintf("%s:%d", ip4, port)
}
