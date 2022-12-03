package client

import (
	"fmt"
	"net/rpc"

	"github.com/eqr/transferit/app/service"
)

type Client struct {
	*rpc.Client
}

func Connect(url string) (*Client, error) {
	client, err := rpc.Dial("tcp", url)
	if err != nil {
		return nil, fmt.Errorf("cannot dial to rpc service: %w", err)
	}

	return &Client{
		Client: client,
	}, nil
}

func (c *Client) Upload(filePath string) error {
	return upload(filePath, c.Client)
}

func (c *Client) Download(id service.TransferID) (string, error) {
	return download(id)
}
