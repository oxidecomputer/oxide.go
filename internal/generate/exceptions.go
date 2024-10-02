// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

// Returns a list of types that should not be omitted when empty
// for json serialisation
func omitemptyExceptions() []string {
	return []string{
		"[]VpcFirewallRuleUpdate",
		"[]NameOrId",
	}
}

func emptyTypes() []string {
	return []string{
		"BgpMessageHistory",
		"SwitchLinkState",
	}
}

// TODO: Actually handle nullable fields properly
func nullable() []string {
	return []string{
		"InstanceDiskAttachment",
	}
}
