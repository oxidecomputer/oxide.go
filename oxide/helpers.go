// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import "fmt"

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
