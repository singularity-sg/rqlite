// Package disco controls interaction with the rqlite Discovery service
package disco

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Response represents the response returned by a Discovery Service.
type Response struct {
	CreatedAt string   `json:"created_at"`
	DiscoID   string   `json:"disco_id"`
	Nodes     []string `json:"nodes"`
}

// Client provides a Discovery Service client.
type Client struct {
	url    string
	logger *log.Logger
}

// New returns an initialized Discovery Service client.
func New(url string) *Client {
	return &Client{
		url:    url,
		logger: log.New(os.Stderr, "[discovery] ", log.LstdFlags),
	}
}

// URL returns the Discovery Service URL used by this client.
func (c *Client) URL() string {
	return c.url
}

// Register attempts to register with the Discovery Service, using the given
// address.
func (c *Client) Register(id, addr string) (*Response, error) {
	m := map[string]string{
		"addr": addr,
	}

	url := c.registrationURL(c.url, id)

	for {
		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}

		c.logger.Println("discovery client attempting registration at", url)
		resp, err := http.Post(url, "application-type/json", bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			r := &Response{}
			if err := json.Unmarshal(b, r); err != nil {
				return nil, err
			}
			return r, nil
		case http.StatusMovedPermanently:
			url = c.registrationURL(resp.Header.Get("location"), id)
			c.logger.Println("discovery client redirecting to", url)
			continue
		default:
			return nil, errors.New(resp.Status)
		}
	}
}

func (c *Client) registrationURL(url, id string) string {
	return fmt.Sprintf("%s/%s", url, id)
}
