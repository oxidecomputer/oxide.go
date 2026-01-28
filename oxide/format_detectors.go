// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import "net"

// formatDetectors maps OpenAPI format strings to detection functions.
// Used by generated code to discriminate untagged union variants.
var formatDetectors = map[string]func(string) bool{
	"ipv4": detectIPv4Format,
	"ipv6": detectIPv6Format,
}

func detectIPv4Format(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

func detectIPv6Format(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() == nil
}
