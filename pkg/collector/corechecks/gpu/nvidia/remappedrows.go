// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build linux && nvml

package nvidia

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"

	"github.com/DataDog/datadog-agent/pkg/metrics"
)

const remappedRowsMetricPrefix = "remapped_rows"

type remappedRowsCollector struct {
	device nvml.Device
}

// newRemappedRowsCollector creates a new remappedRowsMetricsCollector for the given NVML device.
func newRemappedRowsCollector(device nvml.Device) (Collector, error) {
	// Do a first check to see if the device supports remapped rows metrics
	_, _, _, _, ret := device.GetRemappedRows()
	if ret == nvml.ERROR_NOT_SUPPORTED {
		return nil, errUnsupportedDevice
	} else if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("cannot check remapped rows support: %s", nvml.ErrorString(ret))
	}

	return &remappedRowsCollector{
		device: device,
	}, nil
}

func (c *remappedRowsCollector) DeviceUUID() string {
	uuid, _ := c.device.GetUUID()
	return uuid
}

// Collect collects remapped rows metrics from the NVML device.
func (c *remappedRowsCollector) Collect() ([]Metric, error) {
	// Collect remapped rows metrics from the NVML device
	correctable, uncorrectable, pending, failed, ret := c.device.GetRemappedRows()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("cannot get remapped rows: %s", nvml.ErrorString(ret))
	}

	return []Metric{
		{Name: fmt.Sprintf("%s.correctable", remappedRowsMetricPrefix), Value: float64(correctable), Type: metrics.CountType},
		{Name: fmt.Sprintf("%s.uncorrectable", remappedRowsMetricPrefix), Value: float64(uncorrectable), Type: metrics.CountType},
		{Name: fmt.Sprintf("%s.pending", remappedRowsMetricPrefix), Value: boolToFloat(pending), Type: metrics.CountType},
		{Name: fmt.Sprintf("%s.failed", remappedRowsMetricPrefix), Value: boolToFloat(failed), Type: metrics.CountType},
	}, nil
}

// Name returns the name of the collector.
func (c *remappedRowsCollector) Name() CollectorName {
	return remappedRows
}
