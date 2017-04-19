package network

import (
	"net"
	"strings"
)
func GetLocalIP() (address string, err error) {
	interfaces,err := net.Interfaces()
	if err != nil {
		return "",err
	}
	for _,i := range interfaces {
		if (i.Flags & net.FlagLoopback) != 0 {
			continue
		}
		if !((i.Flags & net.FlagUp) != 0) {
			continue
		}
		addr,_ := i.Addrs()
		address := addr[0].String()
		pos := strings.Index(address,"/")
		if pos != -1 {
			address = address[0:pos]
		}
		return address,nil
		
	}
	return "",nil
}