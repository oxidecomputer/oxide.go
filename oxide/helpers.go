// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

// This file contains hand-written helper methods for generated types.

import (
	"encoding/json"
	"fmt"
)

// String helpers for oneOf types whose variants are all string or string-like (Name, NameOrId,
// etc.).

// String returns the string representation of the RouteDestination's value.
// Returns an empty string if no variant is set.
func (v RouteDestination) String() string {
	if v.Value == nil {
		return ""
	}
	switch val := v.Value.(type) {
	case *RouteDestinationIp:
		return val.Value
	case *RouteDestinationIpNet:
		return fmt.Sprintf("%v", val.Value)
	case *RouteDestinationVpc:
		return string(val.Value)
	case *RouteDestinationSubnet:
		return string(val.Value)
	default:
		panic(fmt.Sprintf("unhandled RouteDestination variant: %T", val))
	}
}

// String returns the string representation of the VpcFirewallRuleHostFilter's value.
// Returns an empty string if no variant is set.
func (v VpcFirewallRuleHostFilter) String() string {
	if v.Value == nil {
		return ""
	}
	switch val := v.Value.(type) {
	case *VpcFirewallRuleHostFilterVpc:
		return string(val.Value)
	case *VpcFirewallRuleHostFilterSubnet:
		return string(val.Value)
	case *VpcFirewallRuleHostFilterInstance:
		return string(val.Value)
	case *VpcFirewallRuleHostFilterIp:
		return val.Value
	case *VpcFirewallRuleHostFilterIpNet:
		return fmt.Sprintf("%v", val.Value)
	default:
		panic(fmt.Sprintf("unhandled VpcFirewallRuleHostFilter variant: %T", val))
	}
}

// String returns the string representation of the VpcFirewallRuleTarget's value.
// Returns an empty string if no variant is set.
func (v VpcFirewallRuleTarget) String() string {
	if v.Value == nil {
		return ""
	}
	switch val := v.Value.(type) {
	case *VpcFirewallRuleTargetVpc:
		return string(val.Value)
	case *VpcFirewallRuleTargetSubnet:
		return string(val.Value)
	case *VpcFirewallRuleTargetInstance:
		return string(val.Value)
	case *VpcFirewallRuleTargetIp:
		return val.Value
	case *VpcFirewallRuleTargetIpNet:
		return fmt.Sprintf("%v", val.Value)
	default:
		panic(fmt.Sprintf("unhandled VpcFirewallRuleTarget variant: %T", val))
	}
}

// String returns the string representation of the RouteTarget's value.
// Returns an empty string if no variant is set or for Drop targets.
func (v RouteTarget) String() string {
	if v.Value == nil {
		return ""
	}
	switch val := v.Value.(type) {
	case *RouteTargetIp:
		return val.Value
	case *RouteTargetVpc:
		return string(val.Value)
	case *RouteTargetSubnet:
		return string(val.Value)
	case *RouteTargetInstance:
		return string(val.Value)
	case *RouteTargetInternetGateway:
		return string(val.Value)
	case *RouteTargetDrop:
		return ""
	default:
		panic(fmt.Sprintf("unhandled RouteTarget variant: %T", val))
	}
}

// NewIpNet creates an IpNet from a string value (e.g., "192.168.1.0/24" or "fd00::/64").
// The string is parsed to determine whether it's an IPv4 or IPv6 network.
func NewIpNet(value string) (IpNet, error) {
	var ipNet IpNet
	if err := json.Unmarshal([]byte(`"`+value+`"`), &ipNet); err != nil {
		return IpNet{}, fmt.Errorf("invalid IP network %q: %w", value, err)
	}
	return ipNet, nil
}

// MustIpNet creates an IpNet from a string value, panicking on error.
// Use this only for known-good values.
func MustIpNet(value string) IpNet {
	ipNet, err := NewIpNet(value)
	if err != nil {
		panic(err)
	}
	return ipNet
}

// String returns the string representation of the IpNet.
func (v IpNet) String() string {
	if v.Value == nil {
		return ""
	}
	switch val := v.Value.(type) {
	case *Ipv4Net:
		return string(*val)
	case *Ipv6Net:
		return string(*val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// NewIpRange creates an IpRange from first and last IP strings.
// The IPs are parsed to determine whether they're IPv4 or IPv6.
func NewIpRange(first, last string) (IpRange, error) {
	data := fmt.Sprintf(`{"first":%q,"last":%q}`, first, last)
	var ipRange IpRange
	if err := json.Unmarshal([]byte(data), &ipRange); err != nil {
		return IpRange{}, fmt.Errorf("invalid IP range %q-%q: %w", first, last, err)
	}
	return ipRange, nil
}

// String returns the string representation of the IpRange.
func (v IpRange) String() string {
	if v.Value == nil {
		return ""
	}
	switch val := v.Value.(type) {
	case *Ipv4Range:
		return fmt.Sprintf("%s-%s", val.First, val.Last)
	case *Ipv6Range:
		return fmt.Sprintf("%s-%s", val.First, val.Last)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Constructor helpers for oneOf types whose variants are all string or string-like.

// NewRouteDestination creates a RouteDestination from a type constant and string value.
func NewRouteDestination(t RouteDestinationType, value string) (RouteDestination, error) {
	switch t {
	case RouteDestinationTypeIp:
		return RouteDestination{Value: &RouteDestinationIp{Value: value}}, nil
	case RouteDestinationTypeIpNet:
		ipNet, err := NewIpNet(value)
		if err != nil {
			return RouteDestination{}, err
		}
		return RouteDestination{Value: &RouteDestinationIpNet{Value: ipNet}}, nil
	case RouteDestinationTypeVpc:
		return RouteDestination{Value: &RouteDestinationVpc{Value: Name(value)}}, nil
	case RouteDestinationTypeSubnet:
		return RouteDestination{Value: &RouteDestinationSubnet{Value: Name(value)}}, nil
	default:
		return RouteDestination{}, fmt.Errorf("unknown RouteDestinationType: %s", t)
	}
}

// NewRouteTarget creates a RouteTarget from a type constant and string value.
// For RouteTargetTypeDrop, the value is ignored.
func NewRouteTarget(t RouteTargetType, value string) (RouteTarget, error) {
	switch t {
	case RouteTargetTypeIp:
		return RouteTarget{Value: &RouteTargetIp{Value: value}}, nil
	case RouteTargetTypeVpc:
		return RouteTarget{Value: &RouteTargetVpc{Value: Name(value)}}, nil
	case RouteTargetTypeSubnet:
		return RouteTarget{Value: &RouteTargetSubnet{Value: Name(value)}}, nil
	case RouteTargetTypeInstance:
		return RouteTarget{Value: &RouteTargetInstance{Value: Name(value)}}, nil
	case RouteTargetTypeInternetGateway:
		return RouteTarget{Value: &RouteTargetInternetGateway{Value: Name(value)}}, nil
	case RouteTargetTypeDrop:
		return RouteTarget{Value: &RouteTargetDrop{}}, nil
	default:
		return RouteTarget{}, fmt.Errorf("unknown RouteTargetType: %s", t)
	}
}

// NewVpcFirewallRuleHostFilter creates a VpcFirewallRuleHostFilter from a type
// constant and string value.
func NewVpcFirewallRuleHostFilter(
	t VpcFirewallRuleHostFilterType,
	value string,
) (VpcFirewallRuleHostFilter, error) {
	switch t {
	case VpcFirewallRuleHostFilterTypeVpc:
		return VpcFirewallRuleHostFilter{
			Value: &VpcFirewallRuleHostFilterVpc{Value: Name(value)},
		}, nil
	case VpcFirewallRuleHostFilterTypeSubnet:
		return VpcFirewallRuleHostFilter{
			Value: &VpcFirewallRuleHostFilterSubnet{Value: Name(value)},
		}, nil
	case VpcFirewallRuleHostFilterTypeInstance:
		return VpcFirewallRuleHostFilter{
			Value: &VpcFirewallRuleHostFilterInstance{Value: Name(value)},
		}, nil
	case VpcFirewallRuleHostFilterTypeIp:
		return VpcFirewallRuleHostFilter{Value: &VpcFirewallRuleHostFilterIp{Value: value}}, nil
	case VpcFirewallRuleHostFilterTypeIpNet:
		ipNet, err := NewIpNet(value)
		if err != nil {
			return VpcFirewallRuleHostFilter{}, err
		}
		return VpcFirewallRuleHostFilter{
			Value: &VpcFirewallRuleHostFilterIpNet{Value: ipNet},
		}, nil
	default:
		return VpcFirewallRuleHostFilter{}, fmt.Errorf(
			"unknown VpcFirewallRuleHostFilterType: %s",
			t,
		)
	}
}

// NewVpcFirewallRuleTarget creates a VpcFirewallRuleTarget from a type constant
// and string value.
func NewVpcFirewallRuleTarget(
	t VpcFirewallRuleTargetType,
	value string,
) (VpcFirewallRuleTarget, error) {
	switch t {
	case VpcFirewallRuleTargetTypeVpc:
		return VpcFirewallRuleTarget{Value: &VpcFirewallRuleTargetVpc{Value: Name(value)}}, nil
	case VpcFirewallRuleTargetTypeSubnet:
		return VpcFirewallRuleTarget{Value: &VpcFirewallRuleTargetSubnet{Value: Name(value)}}, nil
	case VpcFirewallRuleTargetTypeInstance:
		return VpcFirewallRuleTarget{Value: &VpcFirewallRuleTargetInstance{Value: Name(value)}}, nil
	case VpcFirewallRuleTargetTypeIp:
		return VpcFirewallRuleTarget{Value: &VpcFirewallRuleTargetIp{Value: value}}, nil
	case VpcFirewallRuleTargetTypeIpNet:
		ipNet, err := NewIpNet(value)
		if err != nil {
			return VpcFirewallRuleTarget{}, err
		}
		return VpcFirewallRuleTarget{Value: &VpcFirewallRuleTargetIpNet{Value: ipNet}}, nil
	default:
		return VpcFirewallRuleTarget{}, fmt.Errorf("unknown VpcFirewallRuleTargetType: %s", t)
	}
}
