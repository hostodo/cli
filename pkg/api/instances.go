package api

import (
	"fmt"
)

// ListInstances retrieves all instances for the authenticated user
func (c *Client) ListInstances(limit, offset int) (*InstancesResponse, error) {
	path := fmt.Sprintf("/client/instances/?limit=%d&offset=%d&sort=updated_at&order=asc", limit, offset)

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	var instancesResp InstancesResponse
	if err := parseResponse(resp, &instancesResp); err != nil {
		return nil, err
	}

	return &instancesResp, nil
}

// GetInstance retrieves details for a specific instance
func (c *Client) GetInstance(instanceID string) (*Instance, error) {
	path := fmt.Sprintf("/client/instances/%s/", instanceID)

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	// Try to parse as wrapped response first
	var wrappedResp InstanceDetailResponse
	if err := parseResponse(resp, &wrappedResp); err == nil && wrappedResp.Instance.InstanceID != "" {
		return &wrappedResp.Instance, nil
	}

	// If that fails, try to parse as direct instance
	var instance Instance
	resp2, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	if err := parseResponse(resp2, &instance); err != nil {
		return nil, err
	}

	return &instance, nil
}

// GetInstancePowerStatus retrieves the power status for an instance
func (c *Client) GetInstancePowerStatus(instanceID string) (string, error) {
	path := fmt.Sprintf("/client/instances/%s/power_status/", instanceID)

	resp, err := c.Get(path)
	if err != nil {
		return "", err
	}

	var statusResp PowerStatusResponse
	if err := parseResponse(resp, &statusResp); err != nil {
		return "", err
	}

	return statusResp.PowerStatus, nil
}

// ControlInstancePower controls the power state of an instance
func (c *Client) ControlInstancePower(instanceID, action string) error {
	path := fmt.Sprintf("/client/instances/%s/power/", instanceID)

	powerReq := PowerControlRequest{
		Action: action,
	}

	resp, err := c.Post(path, powerReq)
	if err != nil {
		return err
	}

	if err := parseResponse(resp, nil); err != nil {
		return err
	}

	return nil
}

// StartInstance starts a stopped instance
func (c *Client) StartInstance(instanceID string) error {
	return c.ControlInstancePower(instanceID, "start")
}

// StopInstance stops a running instance
func (c *Client) StopInstance(instanceID string) error {
	return c.ControlInstancePower(instanceID, "stop")
}

// RebootInstance reboots an instance
func (c *Client) RebootInstance(instanceID string) error {
	return c.ControlInstancePower(instanceID, "reboot")
}
