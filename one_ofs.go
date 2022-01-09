package oxide

// DiskState is the state of a disk.
type DiskState struct {
	State    DiskStateState `json:"state,omitempty" yaml:"state,omitempty"`
	Instance *Instance      `json:"instance,omitempty" yaml:"instance,omitempty"`
}

// RouteDestination is the destination of a route.
type RouteDestination struct {
	Type *RouteDestinationType `json:"type,omitempty" yaml:"type,omitempty"`
	IP   string                `json:"ip,omitempty" yaml:"ip,omitempty"`
	Name string                `json:"name,omitempty" yaml:"name,omitempty"`
}

// RouteTarget is the target of a route.
type RouteTarget struct {
	Type  RouteTargetType `json:"type,omitempty" yaml:"type,omitempty"`
	Value string          `json:"value,omitempty" yaml:"value,omitempty"`
}

// SagaState is the state of a saga.
type SagaState struct {
	State         *SagaStateState `json:"state,omitempty" yaml:"state,omitempty"`
	ErrorInfo     *SagaErrorInfo  `json:"error_info,omitempty" yaml:"error_info,omitempty"`
	ErrorNodeName string          `json:"error_node_name,omitempty" yaml:"error_node_name,omitempty"`
}

// SagaErrorInfo is the error info of a saga.
type SagaErrorInfo struct {
	Error       SagaErrorInfoError `json:"error,omitempty" yaml:"error,omitempty"`
	SourceError interface{}        `json:"source_error,omitempty" yaml:"source_error,omitempty"`
	Message     string             `json:"message,omitempty" yaml:"message,omitempty"`
}

// VPCFirewallRuleTarget is the target of a firewall rule.
type VPCFirewallRuleTarget struct {
	Type VPCFirewallRuleTargetType `json:"type,omitempty" yaml:"type,omitempty"`
	// Value is names must begin with a lower case ASCII letter, be composed exclusively of lowercase ASCII, uppercase ASCII, numbers, and '-', and may not end with a '-'.
	Value Name `json:"value,omitempty" yaml:"value,omitempty"`
}

// VPCFirewallRuleHostFilter is the host filter of a firewall rule.
type VPCFirewallRuleHostFilter struct {
	Type VPCFirewallRuleHostFilterType `json:"type,omitempty" yaml:"type,omitempty"`
	// Value is names must begin with a lower case ASCII letter, be composed exclusively of lowercase ASCII, uppercase ASCII, numbers, and '-', and may not end with a '-'.
	Value Name `json:"value,omitempty" yaml:"value,omitempty"`
}
