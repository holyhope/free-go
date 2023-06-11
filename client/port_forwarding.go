package client

import (
	"fmt"

	"github.com/nikolalohinski/free-go/types"
)

const (
	ErrPortForwardingRuleNotFound = Error("port forwarding rule not found")
)

func (c *client) ListPortForwardingRules() ([]types.PortForwardingRule, error) {
	response, err := c.Get("fw/redir/", c.withSession)
	if err != nil {
		return nil, fmt.Errorf("failed to GET fw/redir/ endpoint: %w", err)
	}

	result := make([]types.PortForwardingRule, 0)
	if err = c.fromGenericResponse(response, &result); err != nil {
		return nil, fmt.Errorf("failed to get port forwarding rules from generic response: %w", err)
	}

	return result, nil
}

func (c *client) GetPortForwardingRule(identifier int64) (rule types.PortForwardingRule, err error) {
	response, err := c.Get(fmt.Sprintf("fw/redir/%d", identifier), c.withSession)
	if err != nil {
		if response != nil && response.ErrorCode == "noent" {
			return rule, ErrPortForwardingRuleNotFound
		}

		return rule, fmt.Errorf("failed to GET fw/redir/%d endpoint: %w", identifier, err)
	}

	if err = c.fromGenericResponse(response, &rule); err != nil {
		return rule, fmt.Errorf("failed to get a port forwarding rule from a generic response: %w", err)
	}

	return rule, nil
}

func (c *client) CreatePortForwardingRule(
	payload types.PortForwardingRulePayload,
) (rule types.PortForwardingRule, err error) {
	response, err := c.Post("fw/redir/", payload, c.withSession)
	if err != nil {
		return rule, fmt.Errorf("failed to POST to fw/redir/ endpoint: %w", err)
	}

	if err = c.fromGenericResponse(response, &rule); err != nil {
		return rule, fmt.Errorf("failed to get a port forwarding rule from a generic response: %w", err)
	}

	return rule, nil
}

func (c *client) DeletePortForwardingRule(identifier int64) error {
	response, err := c.Delete(fmt.Sprintf("fw/redir/%d", identifier), c.withSession)
	if err != nil {
		if response != nil && response.ErrorCode == "noent" {
			return ErrPortForwardingRuleNotFound
		}

		return fmt.Errorf("failed to GET fw/redir/%d endpoint: %w", identifier, err)
	}

	return nil
}