package kerberos

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const realmDomain = "IPA.REDHAT.COM"

type KerberosOutput struct {
	Principal string           `json:"principal"`
	Tickets   []KerberosTicket `json:"tickets"`
}

type KerberosTicket struct {
	Principal string `json:"Principal"`
	Issued    string `json:"Issued"`
	Expires   string `json:"Expires"`
}

func CheckKerberosTicket() error {
	// use the klist command to check if there is a kerberos ticket
	cmd := exec.Command("klist", "--json")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute klist command: %w", err)
	}

	// parse the output JSON
	var kerberosOutput KerberosOutput
	err = json.Unmarshal(output, &kerberosOutput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal kerberos output: %w", err)
	}

	if len(kerberosOutput.Tickets) == 0 {
		return fmt.Errorf("no kerberos ticket found")
	}

	// Check: Username (Principal) in the format of <username>@IPA.REDHAT.COM
	if !strings.Contains(kerberosOutput.Principal, realmDomain) {
		return fmt.Errorf("kerberos ticket principal does not match the expected realm: %s", kerberosOutput.Principal)
	}

	// Check: There is at least one ticket with expiration in the future and principal matches the username
	for _, ticket := range kerberosOutput.Tickets {
		expirationTime, err := time.Parse("20060102150405", ticket.Expires)
		if err != nil {
			return fmt.Errorf("failed to parse expiration time: %w", err)
		}
		if expirationTime.After(time.Now()) && strings.Contains(ticket.Principal, realmDomain) {
			return nil
		}
	}
	return fmt.Errorf("no kerberos ticket found with expiration in the future and principal matching the username")
}
