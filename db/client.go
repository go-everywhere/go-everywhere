//go:build !js

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	etcdClient *clientv3.Client
}

func NewClient(etcdClient *clientv3.Client) *Client {
	return &Client{
		etcdClient: etcdClient,
	}
}

func (c *Client) Put(ctx context.Context, namespace string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	fullKey := fmt.Sprintf("/%s/%s", namespace, key)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = c.etcdClient.Put(ctx, fullKey, string(data))
	return err
}

func (c *Client) Get(ctx context.Context, namespace string, key string) ([]byte, error) {
	fullKey := fmt.Sprintf("/%s/%s", namespace, key)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.etcdClient.Get(ctx, fullKey)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, ErrKeyNotFound
	}

	return resp.Kvs[0].Value, nil
}

func (c *Client) Delete(ctx context.Context, namespace string, key string) (int64, error) {
	fullKey := fmt.Sprintf("/%s/%s", namespace, key)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.etcdClient.Delete(ctx, fullKey)
	if err != nil {
		return 0, err
	}

	return resp.Deleted, nil
}

func (c *Client) GetAll(ctx context.Context, namespace string) (map[string][]byte, error) {
	prefix := fmt.Sprintf("/%s/", namespace)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.etcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for _, kv := range resp.Kvs {
		// Remove the namespace prefix from the key
		key := string(kv.Key)[len(prefix):]
		result[key] = kv.Value
	}

	return result, nil
}

func (c *Client) Close() error {
	return c.etcdClient.Close()
}

func InitializeNamespaces(client *Client) {
	// Pre-create namespaces if needed (etcd doesn't require this, but keeping for compatibility)
	log.Println("[INFO] etcd client initialized with namespaces support")
}