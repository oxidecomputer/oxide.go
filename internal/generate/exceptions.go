// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

// omitzeroTypes returns a slice of types that should be tagged with omitzero
// for serialization and deserialization.
func omitzeroTypes() []string {
	return []string{
		"[]VpcFirewallRuleUpdate",
		"[]NameOrId",
		"DerEncodedKeyPair",
	}
}

func emptyTypes() []string {
	return []string{
		"BgpMessageHistory",
		"SwitchLinkState",
	}
}

func nullable() []string {
	// TODO: This type has a nested required "Type" field, which hinders
	// the usage of this type. Remove when this is fixed in the upstream API
	return []string{
		"InstanceDiskAttachment",
	}
}
