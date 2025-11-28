package vpn

import (
	"fmt"
	"net"
)

// Check if CLI is connected to VPN attempting to resolve engineering.redhat.com
func CheckConnected() error {
	_, err := net.LookupHost("pyxis.engineering.redhat.com")
	if err != nil {
		return fmt.Errorf("failed to resolve pyxis.engineering.redhat.com: %s. Make sure you are connected to the VPN", err.Error())
	}
	return nil
}
