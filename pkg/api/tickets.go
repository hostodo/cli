package api

import (
	"fmt"
	"net/url"
)

// ListTickets retrieves support tickets with an optional status filter.
func (c *Client) ListTickets(status string) ([]Ticket, error) {
	path := "/v1/helpdesk/tickets/"
	if status != "" {
		path = fmt.Sprintf("%s?status=%s", path, url.QueryEscape(status))
	}

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	var ticketsResp TicketsResponse
	if err := parseResponse(resp, &ticketsResp); err != nil {
		return nil, err
	}

	return ticketsResp.Results, nil
}

// GetTicket retrieves a single support ticket by numeric ID or public ticket ID.
func (c *Client) GetTicket(ticketID string) (*Ticket, error) {
	path := fmt.Sprintf("/v1/helpdesk/tickets/%s/", url.PathEscape(ticketID))

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	var ticket Ticket
	if err := parseResponse(resp, &ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

// CreateTicket opens a new support ticket.
func (c *Client) CreateTicket(req TicketCreateRequest) (*Ticket, error) {
	resp, err := c.Post("/v1/helpdesk/tickets/", req)
	if err != nil {
		return nil, err
	}

	var ticket Ticket
	if err := parseResponse(resp, &ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

// ReplyToTicket adds a reply to an existing support ticket.
func (c *Client) ReplyToTicket(ticketID string, req TicketReplyRequest) (*TicketReply, error) {
	path := fmt.Sprintf("/v1/helpdesk/tickets/%s/reply/", url.PathEscape(ticketID))

	resp, err := c.Post(path, req)
	if err != nil {
		return nil, err
	}

	var reply TicketReply
	if err := parseResponse(resp, &reply); err != nil {
		return nil, err
	}

	return &reply, nil
}

// ListDepartments retrieves enabled helpdesk departments.
func (c *Client) ListDepartments() ([]Department, error) {
	resp, err := c.Get("/v1/helpdesk/departments/")
	if err != nil {
		return nil, err
	}

	var departmentsResp DepartmentsResponse
	if err := parseResponse(resp, &departmentsResp); err != nil {
		return nil, err
	}

	return departmentsResp.Results, nil
}
