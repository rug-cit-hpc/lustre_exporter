// (C) Copyright 2017 Hewlett Packard Enterprise Development LP
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sources

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// string mappings for 'health_check' values
	healthCheckHealthy   string = "1"
	healthCheckUnhealthy string = "0"
)

var (
	// HealthStatusEnabled specifies whether to collect Health metrics
	HealthStatusEnabled string
	// MdtEnabled specifies whether to collect MDT metrics
	MdtEnabled string
	// MgsEnabled specifies whether to collect MGS metrics
	MgsEnabled string
)

func init() {
	Factories["sysfs"] = newLustreSysFsSource
}

type LustreSysFsSource struct {
	lustreProcMetrics []lustreProcMetric
	basePath          string
}

func (s *LustreSysFsSource) generateHealthStatusTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"": {
			{"health_check", "health_check", "Current health status for the indicated instance: " + healthCheckHealthy + " refers to 'healthy', " + healthCheckUnhealthy + " refers to 'unhealthy'", gaugeMetric, false, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "health", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateOSTMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"obdfilter/*-OST*": {
			{"degraded", "degraded", "Binary indicator as to whether or not the pool is degraded - 0 for not degraded, 1 for degraded", gaugeMetric, false, core},
			{"grant_precreate", "grant_precreate_capacity_bytes", "Maximum space in bytes that clients can preallocate for objects", gaugeMetric, false, extended},
			{"grant_compat_disable", "grant_compat_disabled", "Binary indicator as to whether clients with OBD_CONNECT_GRANT_PARAM setting will be granted space", gaugeMetric, false, extended},
			{"job_cleanup_interval", "job_cleanup_interval_seconds", "Interval in seconds between cleanup of tuning statistics", gaugeMetric, false, extended},
			{"lfsck_speed_limit", "lfsck_speed_limit", "Maximum operations per second LFSCK (Lustre filesystem verification) can run", gaugeMetric, false, extended},
			{"num_exports", "exports_total", "Total number of times the pool has been exported", counterMetric, false, core},
			{"recovery_time_hard", "recovery_time_hard_seconds", "Maximum timeout 'recover_time_soft' can increment to for a single server", gaugeMetric, false, extended},
			{"recovery_time_soft", "recovery_time_soft_seconds", "Duration in seconds for a client to attempt to reconnect after a crash (automatically incremented if servers are still in an error state)", gaugeMetric, false, extended},
			{"precreate_batch", "precreate_batch", "Maximum number of objects that can be included in a single transaction", gaugeMetric, false, extended},
			{"soft_sync_limit", "soft_sync_limit", "Number of RPCs necessary before triggering a sync", gaugeMetric, false, extended},
			{"sync_journal", "sync_journal_enabled", "Binary indicator as to whether or not the journal is set for asynchronous commits", gaugeMetric, false, extended},
			{"tot_dirty", "exports_dirty_total", "Total number of exports that have been marked dirty", counterMetric, false, core},
			{"tot_granted", "exports_granted_total", "Total number of exports that have been marked granted", counterMetric, false, core},
			{"tot_pending", "exports_pending_total", "Total number of exports that have been marked pending", counterMetric, false, core},
		},
		"osd-*/*-OST*": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", gaugeMetric, false, core},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", gaugeMetric, false, core},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gaugeMetric, false, core},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes free in the pool", gaugeMetric, false, core},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", gaugeMetric, false, core},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", gaugeMetric, false, core},
		},
		"ldlm/namespaces/filter-*": {
			{"lock_count", "lock_count", "Number of locks", gaugeMetric, false, extended},
			{"lock_timeouts", "lock_timeout", "Number of lock timeouts", counterMetric, false, extended},
			{"contended_locks", "lock_contended", "Number of contended locks", gaugeMetric, false, extended},
			{"contention_seconds", "lock_contention_seconds", "Time in seconds during which locks were contended", gaugeMetric, false, extended},

			{"pool/granted", "lock_granted", "Number of granted locks", gaugeMetric, false, extended},
			{"pool/grant_plan", "lock_grant_plan", "Number of planned lock grants per second", gaugeMetric, false, extended},
			{"pool/grant_rate", "lock_grant_rate", "Lock grant rate", gaugeMetric, false, extended},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "ost", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateMDTMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"osd-*/*-MDT*": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", gaugeMetric, false, core},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", gaugeMetric, false, core},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gaugeMetric, false, core},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", gaugeMetric, false, core},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes free in the pool", gaugeMetric, false, core},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", gaugeMetric, false, core},
		},
		"mdt/*": {
			{"num_exports", "exports_total", "Total number of times the pool has been exported", counterMetric, false, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "mdt", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateMGSMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"mgs/MGS/osd/": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", gaugeMetric, false, core},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", gaugeMetric, false, core},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gaugeMetric, false, core},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", gaugeMetric, false, core},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes free in the pool", gaugeMetric, false, core},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", gaugeMetric, false, core},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "mgs", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func (s *LustreSysFsSource) generateClientMetricTemplates(filter string) {
	metricMap := map[string][]lustreHelpStruct{
		"llite/*": {
			{"blocksize", "blocksize_bytes", "Filesystem block size in bytes", gaugeMetric, false, core},
			{"checksum_pages", "checksum_pages_enabled", "Returns '1' if data checksumming is enabled for the client", gaugeMetric, false, extended},
			{"default_easize", "default_ea_size_bytes", "Default Extended Attribute (EA) size in bytes", gaugeMetric, false, extended},
			{"filesfree", "inodes_free", "The number of inodes (objects) available", gaugeMetric, false, core},
			{"filestotal", "inodes_maximum", "The maximum number of inodes (objects) the filesystem can hold", gaugeMetric, false, core},
			{"kbytesavail", "available_kilobytes", "Number of kilobytes readily available in the pool", gaugeMetric, false, core},
			{"kbytesfree", "free_kilobytes", "Number of kilobytes free in the pool", gaugeMetric, false, core},
			{"kbytestotal", "capacity_kilobytes", "Capacity of the pool in kilobytes", gaugeMetric, false, core},
			{"lazystatfs", "lazystatfs_enabled", "Returns '1' if lazystatfs (a non-blocking alternative to statfs) is enabled for the client", gaugeMetric, false, extended},
			{"max_easize", "maximum_ea_size_bytes", "Maximum Extended Attribute (EA) size in bytes", gaugeMetric, false, extended},
			{"max_read_ahead_mb", "maximum_read_ahead_megabytes", "Maximum number of megabytes to read ahead", gaugeMetric, false, extended},
			{"max_read_ahead_per_file_mb", "maximum_read_ahead_per_file_megabytes", "Maximum number of megabytes per file to read ahead", gaugeMetric, false, extended},
			{"max_read_ahead_whole_mb", "maximum_read_ahead_whole_megabytes", "Maximum file size in megabytes for a file to be read in its entirety", gaugeMetric, false, extended},
			{"statahead_agl", "statahead_agl_enabled", "Returns '1' if the Asynchronous Glimpse Lock (AGL) for statahead is enabled", gaugeMetric, false, extended},
			{"statahead_max", "statahead_maximum", "Maximum window size for statahead", gaugeMetric, false, extended},
			{"stats", "read_samples_total", readSamplesHelp, counterMetric, false, core},
			{"stats", "read_minimum_size_bytes", readMinimumHelp, gaugeMetric, false, extended},
			{"stats", "read_maximum_size_bytes", readMaximumHelp, gaugeMetric, false, extended},
			{"stats", "read_bytes_total", readTotalHelp, counterMetric, false, core},
			{"stats", "write_samples_total", writeSamplesHelp, counterMetric, false, core},
			{"stats", "write_minimum_size_bytes", writeMinimumHelp, gaugeMetric, false, extended},
			{"stats", "write_maximum_size_bytes", writeMaximumHelp, gaugeMetric, false, extended},
			{"stats", "write_bytes_total", writeTotalHelp, counterMetric, false, core},
			{"stats", "stats_total", statsHelp, counterMetric, true, core},
			{"xattr_cache", "xattr_cache_enabled", "Returns '1' if extended attribute cache is enabled", gaugeMetric, false, extended},
		},
	}
	for path := range metricMap {
		for _, item := range metricMap[path] {
			if filter == extended || item.priorityLevel == core {
				newMetric := newLustreProcMetric(item.filename, item.promName, "client", path, item.helpText, item.hasMultipleVals, item.metricFunc)
				s.lustreProcMetrics = append(s.lustreProcMetrics, *newMetric)
			}
		}
	}
}

func newLustreSysFsSource() LustreSource {
	var l LustreSysFsSource
	l.basePath = filepath.Join(SysLocation, "fs/lustre")
	if HealthStatusEnabled != disabled {
		l.generateHealthStatusTemplates(HealthStatusEnabled)
	}
	if OstEnabled != disabled {
		l.generateOSTMetricTemplates(OstEnabled)
	}
	if MdtEnabled != disabled {
		l.generateMDTMetricTemplates(MdtEnabled)
	}
	if MgsEnabled != disabled {
		l.generateMGSMetricTemplates(MgsEnabled)
	}
	if ClientEnabled != disabled {
		l.generateClientMetricTemplates(ClientEnabled)
	}
	return &l
}

func (s *LustreSysFsSource) Update(ch chan<- prometheus.Metric) (err error) {
	var metricType string
	var directoryDepth int

	for _, metric := range s.lustreProcMetrics {
		directoryDepth = strings.Count(metric.filename, "/")
		paths, err := filepath.Glob(filepath.Join(s.basePath, metric.path, metric.filename))
		if err != nil {
			return err
		}
		if paths == nil {
			continue
		}
		for _, path := range paths {
			metricType = single
			switch metric.filename {
			case "health_check":
				err = s.parseTextFile(metric.source, "health_check", path, directoryDepth, metric.helpText, metric.promName, func(nodeType string, nodeName string, name string, helpText string, value float64) {
					ch <- metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value)
				})
				if err != nil {
					return err
				}
			default:
				if metric.filename == stats {
					metricType = stats
				} else if metric.filename == mdStats {
					metricType = mdStats
				}
				err = s.parseFile(metric.source, metricType, path, directoryDepth, metric.helpText, metric.promName, metric.hasMultipleVals, func(nodeType string, nodeName string, name string, helpText string, value float64, extraLabel string, extraLabelValue string) {
					if extraLabelValue == "" {
						ch <- metric.metricFunc([]string{"component", "target"}, []string{nodeType, nodeName}, name, helpText, value)
					} else {
						ch <- metric.metricFunc([]string{"component", "target", extraLabel}, []string{nodeType, nodeName, extraLabelValue}, name, helpText, value)
					}
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *LustreSysFsSource) parseTextFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, handler func(string, string, string, string, float64)) (err error) {
	filename, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	fileBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	fileString := string(fileBytes[:])
	switch filename {
	case "health_check":
		if strings.TrimSpace(fileString) == "healthy" {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckHealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		} else {
			value, err := strconv.ParseFloat(strings.TrimSpace(healthCheckUnhealthy), 64)
			if err != nil {
				return err
			}
			handler(nodeType, nodeName, promName, helpText, value)
		}
	}
	return nil
}

func (s *LustreSysFsSource) parseFile(nodeType string, metricType string, path string, directoryDepth int, helpText string, promName string, hasMultipleVals bool, handler func(string, string, string, string, float64, string, string)) (err error) {
	_, nodeName, err := parseFileElements(path, directoryDepth)
	if err != nil {
		return err
	}
	switch metricType {
	case single:
		value, err := ioutil.ReadFile(filepath.Clean(path))
		if err != nil {
			return err
		}
		convertedValue, err := strconv.ParseFloat(strings.TrimSpace(string(value)), 64)
		if err != nil {
			return err
		}
		handler(nodeType, nodeName, promName, helpText, convertedValue, "", "")
	case stats, mdStats:
		metricList, err := parseStatsFile(helpText, promName, path, hasMultipleVals)
		if err != nil {
			return err
		}

		for _, metric := range metricList {
			handler(nodeType, nodeName, metric.title, metric.help, metric.value, metric.extraLabel, metric.extraLabelValue)
		}
	}
	return nil
}
