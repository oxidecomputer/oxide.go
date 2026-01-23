// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

// This file contains hand-written helper methods for generated types.

import "fmt"

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

// Constructor helpers for oneOf types whose variants are all string or string-like.

// NewRouteDestination creates a RouteDestination from a type constant and string value.
func NewRouteDestination(t RouteDestinationType, value string) (RouteDestination, error) {
	switch t {
	case RouteDestinationTypeIp:
		return RouteDestination{Value: &RouteDestinationIp{Value: value}}, nil
	case RouteDestinationTypeIpNet:
		return RouteDestination{Value: &RouteDestinationIpNet{Value: IpNet(value)}}, nil
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
		return VpcFirewallRuleHostFilter{
			Value: &VpcFirewallRuleHostFilterIpNet{Value: IpNet(value)},
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
		return VpcFirewallRuleTarget{Value: &VpcFirewallRuleTargetIpNet{Value: IpNet(value)}}, nil
	default:
		return VpcFirewallRuleTarget{}, fmt.Errorf("unknown VpcFirewallRuleTargetType: %s", t)
	}
}
