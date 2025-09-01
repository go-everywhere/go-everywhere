//go:build !js

package db

import (
	"context"
	"log"

	"github.com/olric-data/olric"
)

type Client struct {
	embedded *olric.EmbeddedClient
	dmaps    map[string]olric.DMap
}

func NewClient(embedded *olric.EmbeddedClient) *Client {
	return &Client{
		embedded: embedded,
		dmaps:    make(map[string]olric.DMap),
	}
}

func (c *Client) GetDMap(name string) (olric.DMap, error) {
	if dm, exists := c.dmaps[name]; exists {
		return dm, nil
	}
	
	dm, err := c.embedded.NewDMap(name)
	if err != nil {
		return nil, err
	}
	
	c.dmaps[name] = dm
	return dm, nil
}

func (c *Client) Users() (olric.DMap, error) {
	return c.GetDMap("users")
}

func (c *Client) Put(ctx context.Context, dmap string, key string, value interface{}) error {
	dm, err := c.GetDMap(dmap)
	if err != nil {
		return err
	}
	return dm.Put(ctx, key, value)
}

func (c *Client) Get(ctx context.Context, dmap string, key string) (*olric.GetResponse, error) {
	dm, err := c.GetDMap(dmap)
	if err != nil {
		return nil, err
	}
	return dm.Get(ctx, key)
}

func (c *Client) Delete(ctx context.Context, dmap string, key string) (int, error) {
	dm, err := c.GetDMap(dmap)
	if err != nil {
		return 0, err
	}
	return dm.Delete(ctx, key)
}

func InitializeDMaps(client *olric.EmbeddedClient) {
	c := NewClient(client)
	
	// Pre-create DMaps
	if _, err := c.Users(); err != nil {
		log.Printf("Warning: Failed to create users DMap: %v", err)
	}
	
	// Add other DMaps here as needed
}