package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

type Client struct {
	sockPath string
}

func NewClient(sockPath string) *Client {
	return &Client{sockPath: sockPath}
}

func (c *Client) send(req Request) (*Response, error) {
	conn, err := net.Dial("unix", c.sockPath)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to daemon (is it running?): %w", err)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, err
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) Ping() error {
	resp, err := c.send(Request{Action: ActionPing})
	if err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Message)
	}
	return nil
}

func (c *Client) Add(domain string, port int, dir string, https bool) error {
	resp, err := c.send(Request{Action: ActionAdd, Domain: domain, Port: port, Dir: dir, HTTPS: https})
	if err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Message)
	}
	return nil
}

func (c *Client) Remove(domain string) error {
	resp, err := c.send(Request{Action: ActionRemove, Domain: domain})
	if err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Message)
	}
	return nil
}

func (c *Client) List() ([]ProjectStatus, error) {
	resp, err := c.send(Request{Action: ActionList})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Message)
	}
	return resp.Projects, nil
}

func (c *Client) Status() (*DaemonStatus, error) {
	resp, err := c.send(Request{Action: ActionStatus})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Message)
	}
	return resp.Status, nil
}

func (c *Client) StopDaemon() error {
	resp, err := c.send(Request{Action: ActionStop})
	if err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Message)
	}
	return nil
}
