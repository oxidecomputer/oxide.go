// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import "net"

// Format detection functions for untagged union discrimination.
// Generated code calls these directly, so missing detectors cause compile errors.

// DetectIpv4Format returns true if s is a valid IPv4 address.
func DetectIpv4Format(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

// DetectIpv6Format returns true if s is a valid IPv6 address.
func DetectIpv6Format(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() == nil
}
