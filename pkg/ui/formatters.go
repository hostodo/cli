package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hostodo/hostodo-cli/pkg/api"
)

// FormatInstancesJSON formats instances as JSON
func FormatInstancesJSON(instances []api.Instance) (string, error) {
	data, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// FormatInstancesSimpleTable formats instances as a simple ASCII table
func FormatInstancesSimpleTable(instances []api.Instance) string {
	if len(instances) == 0 {
		return "No instances found"
	}

	// Define column widths
	const (
		idWidth       = 12
		hostnameWidth = 25
		ipWidth       = 16
		statusWidth   = 14
		powerWidth    = 10
		ramWidth      = 8
		cpuWidth      = 6
		diskWidth     = 8
	)

	var sb strings.Builder

	// Header
	header := fmt.Sprintf(
		"%-*s  %-*s  %-*s  %-*s  %-*s  %*s  %*s  %*s",
		idWidth, "ID",
		hostnameWidth, "HOSTNAME",
		ipWidth, "IP ADDRESS",
		statusWidth, "STATUS",
		powerWidth, "POWER",
		ramWidth, "RAM (MB)",
		cpuWidth, "CPU",
		diskWidth, "DISK (GB)",
	)
	sb.WriteString(header + "\n")
	sb.WriteString(strings.Repeat("-", len(header)) + "\n")

	// Rows
	for _, instance := range instances {
		row := fmt.Sprintf(
			"%-*s  %-*s  %-*s  %-*s  %-*s  %*d  %*d  %*d",
			idWidth, truncate(instance.InstanceID, idWidth),
			hostnameWidth, truncate(instance.Hostname, hostnameWidth),
			ipWidth, truncate(instance.MainIP, ipWidth),
			statusWidth, truncate(instance.Status, statusWidth),
			powerWidth, truncate(instance.PowerStatus, powerWidth),
			ramWidth, instance.RAM,
			cpuWidth, instance.VCPU,
			diskWidth, instance.Disk,
		)
		sb.WriteString(row + "\n")
	}

	return sb.String()
}

// FormatInstancesDetailedTable formats instances with more details
func FormatInstancesDetailedTable(instances []api.Instance) string {
	if len(instances) == 0 {
		return "No instances found"
	}

	var sb strings.Builder

	for i, instance := range instances {
		if i > 0 {
			sb.WriteString("\n")
		}

		sb.WriteString(fmt.Sprintf("Instance: %s\n", instance.InstanceID))
		sb.WriteString(fmt.Sprintf("  Hostname:     %s\n", instance.Hostname))
		sb.WriteString(fmt.Sprintf("  IP Address:   %s\n", instance.MainIP))
		if len(instance.IPs) > 1 {
			sb.WriteString(fmt.Sprintf("  Additional:   %s\n", strings.Join(instance.IPs[1:], ", ")))
		}
		sb.WriteString(fmt.Sprintf("  Status:       %s\n", instance.Status))
		sb.WriteString(fmt.Sprintf("  Power:        %s\n", instance.PowerStatus))
		sb.WriteString(fmt.Sprintf("  Resources:    %d MB RAM, %d CPU, %d GB Disk\n",
			instance.RAM, instance.VCPU, instance.Disk))
		sb.WriteString(fmt.Sprintf("  Bandwidth:    %.2f / %d GB\n",
			instance.BandwidthUsage, instance.Bandwidth))
		sb.WriteString(fmt.Sprintf("  Plan:         %s\n", instance.Plan.Name))
		sb.WriteString(fmt.Sprintf("  Template:     %s\n", instance.Template.Name))
		sb.WriteString(fmt.Sprintf("  Region:       %s\n", instance.Node.Region))
		sb.WriteString(fmt.Sprintf("  Billing:      $%s / %s\n",
			instance.BillingAmount, instance.BillingCycle))
		sb.WriteString(fmt.Sprintf("  Next Due:     %s\n", instance.NextDueDate))
		if instance.IsSuspended {
			sb.WriteString(fmt.Sprintf("  Suspended:    Yes (%s)\n", instance.SuspensionReason))
		}
	}

	return sb.String()
}

// FormatInstanceDetail formats a single instance with full details
func FormatInstanceDetail(instance *api.Instance) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(TitleStyle.Render("Instance Details") + "\n\n")

	// Basic Info
	sb.WriteString(HeaderStyle.Render("Basic Information") + "\n")
	sb.WriteString(fmt.Sprintf("  ID:           %s\n", instance.InstanceID))
	sb.WriteString(fmt.Sprintf("  Hostname:     %s\n", instance.Hostname))
	sb.WriteString(fmt.Sprintf("  Status:       %s\n", GetPowerStatusBadge(instance.Status)))
	sb.WriteString(fmt.Sprintf("  Power:        %s\n", GetPowerStatusBadge(instance.PowerStatus)))
	sb.WriteString("\n")

	// Network
	sb.WriteString(HeaderStyle.Render("Network") + "\n")
	sb.WriteString(fmt.Sprintf("  Main IP:      %s\n", instance.MainIP))
	if len(instance.IPs) > 1 {
		sb.WriteString(fmt.Sprintf("  Additional:   %s\n", strings.Join(instance.IPs[1:], ", ")))
	}
	sb.WriteString(fmt.Sprintf("  MAC Address:  %s\n", instance.MAC))
	sb.WriteString("\n")

	// Resources
	sb.WriteString(HeaderStyle.Render("Resources") + "\n")
	sb.WriteString(fmt.Sprintf("  RAM:          %d MB\n", instance.RAM))
	sb.WriteString(fmt.Sprintf("  CPU:          %d cores\n", instance.VCPU))
	sb.WriteString(fmt.Sprintf("  Disk:         %d GB\n", instance.Disk))
	sb.WriteString(fmt.Sprintf("  Bandwidth:    %.2f / %d GB (%.1f%%)\n",
		instance.BandwidthUsage, instance.Bandwidth,
		(instance.BandwidthUsage/float64(instance.Bandwidth))*100))
	sb.WriteString("\n")

	// Plan & Template
	sb.WriteString(HeaderStyle.Render("Configuration") + "\n")
	sb.WriteString(fmt.Sprintf("  Plan:         %s\n", instance.Plan.Name))
	sb.WriteString(fmt.Sprintf("  Template:     %s\n", instance.Template.Name))
	sb.WriteString(fmt.Sprintf("  Region:       %s\n", instance.Node.Region))
	sb.WriteString(fmt.Sprintf("  Node:         %s\n", instance.Node.Name))
	sb.WriteString("\n")

	// Billing
	sb.WriteString(HeaderStyle.Render("Billing") + "\n")
	sb.WriteString(fmt.Sprintf("  Amount:       $%s / %s\n", instance.BillingAmount, instance.BillingCycle))
	sb.WriteString(fmt.Sprintf("  Next Due:     %s\n", instance.NextDueDate))
	sb.WriteString(fmt.Sprintf("  Auto-Renew:   %t\n", instance.AutorenewalEnabled))
	if instance.IsSuspended {
		sb.WriteString(fmt.Sprintf("  Suspended:    %s\n", ErrorStyle.Render("Yes - "+instance.SuspensionReason)))
	}
	sb.WriteString("\n")

	// Timestamps
	sb.WriteString(HeaderStyle.Render("Timeline") + "\n")
	sb.WriteString(fmt.Sprintf("  Created:      %s\n", instance.CreatedAt))
	sb.WriteString(fmt.Sprintf("  Updated:      %s\n", instance.UpdatedAt))

	return sb.String()
}

// truncate truncates a string to the specified length
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}
