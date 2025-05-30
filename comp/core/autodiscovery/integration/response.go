// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package integration

// ConfigResponse holds information about the config
// the instance IDs are precomputed to avoid discrepancies between the server and the client
// The InstanceIDs must have the same order as the instances in the Config struct
type ConfigResponse struct {
	InstanceIDs []string `json:"instance_ids"`
	Config      Config   `json:"config"`
}

// ConfigCheckResponse holds the config check response
type ConfigCheckResponse struct {
	Configs         []ConfigResponse    `json:"configs"`
	ResolveWarnings map[string][]string `json:"resolve_warnings"`
	ConfigErrors    map[string]string   `json:"config_errors"`
	Unresolved      map[string]Config   `json:"unresolved"`
}
