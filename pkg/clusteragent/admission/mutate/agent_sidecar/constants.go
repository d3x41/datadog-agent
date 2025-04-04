// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

package agentsidecar

const (
	agentSidecarContainerName = "datadog-agent-injected"
	providerFargate           = "fargate"
)

const (
	agentConfigVolumeName  = "agent-config"
	agentOptionsVolumeName = "agent-option"
	agentTmpVolumeName     = "agent-tmp"
	agentLogsVolumeName    = "agent-log"
)
