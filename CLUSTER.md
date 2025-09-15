# Clustering Guide for Go-Everywhere with Embedded etcd

This guide explains how to deploy multiple instances of the application in a cluster configuration using the embedded etcd for distributed data storage and coordination.

## Overview

The application uses embedded etcd to provide distributed key-value storage across multiple instances. When running in a cluster, all instances share the same data and can handle requests independently while maintaining consistency.

## Prerequisites

- Go 1.24.1 or later
- Network connectivity between cluster nodes
- Ports 2379 (client) and 2380 (peer) available on each node

## Configuration Requirements

To run multiple instances in a cluster, you'll need to modify the `database.go` file to accept configuration from environment variables or command-line flags. Here's what needs to be configured for each instance:

### Required Configuration per Node

1. **Node Name**: Unique identifier for each instance
2. **Client URLs**: Where clients connect to this node
3. **Peer URLs**: Where other etcd nodes connect for replication
4. **Initial Cluster**: List of all initial cluster members
5. **Data Directory**: Unique data directory per instance

## Deployment Scenarios

### Scenario 1: Local Development Cluster (Same Machine)

For testing on a single machine with different ports:

#### Instance 1
```bash
# Set environment variables
export ETCD_NAME=node1
export ETCD_DATA_DIR=/tmp/etcd-node1
export ETCD_CLIENT_PORT=2379
export ETCD_PEER_PORT=2380
export ETCD_INITIAL_CLUSTER="node1=http://127.0.0.1:2380,node2=http://127.0.0.1:2381,node3=http://127.0.0.1:2382"

# Run the application on port 8000
go run .
```

#### Instance 2
```bash
export ETCD_NAME=node2
export ETCD_DATA_DIR=/tmp/etcd-node2
export ETCD_CLIENT_PORT=2389
export ETCD_PEER_PORT=2381
export ETCD_INITIAL_CLUSTER="node1=http://127.0.0.1:2380,node2=http://127.0.0.1:2381,node3=http://127.0.0.1:2382"

# Run on different HTTP port
PORT=8001 go run .
```

#### Instance 3
```bash
export ETCD_NAME=node3
export ETCD_DATA_DIR=/tmp/etcd-node3
export ETCD_CLIENT_PORT=2399
export ETCD_PEER_PORT=2382
export ETCD_INITIAL_CLUSTER="node1=http://127.0.0.1:2380,node2=http://127.0.0.1:2381,node3=http://127.0.0.1:2382"

# Run on different HTTP port
PORT=8002 go run .
```

### Scenario 2: Multi-Server Production Cluster

For production deployment across multiple servers:

#### Server 1 (IP: 10.0.1.10)
```bash
export ETCD_NAME=prod-node1
export ETCD_DATA_DIR=/var/lib/etcd-data
export ETCD_CLIENT_URLS="http://10.0.1.10:2379"
export ETCD_PEER_URLS="http://10.0.1.10:2380"
export ETCD_INITIAL_CLUSTER="prod-node1=http://10.0.1.10:2380,prod-node2=http://10.0.1.11:2380,prod-node3=http://10.0.1.12:2380"
export ETCD_INITIAL_CLUSTER_STATE=new

./your-app-binary
```

#### Server 2 (IP: 10.0.1.11)
```bash
export ETCD_NAME=prod-node2
export ETCD_DATA_DIR=/var/lib/etcd-data
export ETCD_CLIENT_URLS="http://10.0.1.11:2379"
export ETCD_PEER_URLS="http://10.0.1.11:2380"
export ETCD_INITIAL_CLUSTER="prod-node1=http://10.0.1.10:2380,prod-node2=http://10.0.1.11:2380,prod-node3=http://10.0.1.12:2380"
export ETCD_INITIAL_CLUSTER_STATE=new

./your-app-binary
```

#### Server 3 (IP: 10.0.1.12)
```bash
export ETCD_NAME=prod-node3
export ETCD_DATA_DIR=/var/lib/etcd-data
export ETCD_CLIENT_URLS="http://10.0.1.12:2379"
export ETCD_PEER_URLS="http://10.0.1.12:2380"
export ETCD_INITIAL_CLUSTER="prod-node1=http://10.0.1.10:2380,prod-node2=http://10.0.1.11:2380,prod-node3=http://10.0.1.12:2380"
export ETCD_INITIAL_CLUSTER_STATE=new

./your-app-binary
```

## Docker Deployment

### Docker Compose Example

Create a `docker-compose.cluster.yml`:

```yaml
version: '3.8'

services:
  app-node1:
    build: .
    environment:
      - ETCD_NAME=node1
      - ETCD_DATA_DIR=/data/etcd
      - ETCD_CLIENT_URLS=http://app-node1:2379
      - ETCD_PEER_URLS=http://app-node1:2380
      - ETCD_INITIAL_CLUSTER=node1=http://app-node1:2380,node2=http://app-node2:2380,node3=http://app-node3:2380
      - ETCD_INITIAL_CLUSTER_STATE=new
    ports:
      - "8001:8000"
    volumes:
      - node1-data:/data/etcd
    networks:
      - cluster-net

  app-node2:
    build: .
    environment:
      - ETCD_NAME=node2
      - ETCD_DATA_DIR=/data/etcd
      - ETCD_CLIENT_URLS=http://app-node2:2379
      - ETCD_PEER_URLS=http://app-node2:2380
      - ETCD_INITIAL_CLUSTER=node1=http://app-node1:2380,node2=http://app-node2:2380,node3=http://app-node3:2380
      - ETCD_INITIAL_CLUSTER_STATE=new
    ports:
      - "8002:8000"
    volumes:
      - node2-data:/data/etcd
    networks:
      - cluster-net

  app-node3:
    build: .
    environment:
      - ETCD_NAME=node3
      - ETCD_DATA_DIR=/data/etcd
      - ETCD_CLIENT_URLS=http://app-node3:2379
      - ETCD_PEER_URLS=http://app-node3:2380
      - ETCD_INITIAL_CLUSTER=node1=http://app-node1:2380,node2=http://app-node2:2380,node3=http://app-node3:2380
      - ETCD_INITIAL_CLUSTER_STATE=new
    ports:
      - "8003:8000"
    volumes:
      - node3-data:/data/etcd
    networks:
      - cluster-net

volumes:
  node1-data:
  node2-data:
  node3-data:

networks:
  cluster-net:
    driver: bridge
```

Run with:
```bash
docker-compose -f docker-compose.cluster.yml up
```

## Kubernetes Deployment

### StatefulSet Example

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: go-everywhere-cluster
spec:
  serviceName: go-everywhere
  replicas: 3
  selector:
    matchLabels:
      app: go-everywhere
  template:
    metadata:
      labels:
        app: go-everywhere
    spec:
      containers:
      - name: go-everywhere
        image: your-registry/go-everywhere:latest
        ports:
        - containerPort: 8000
          name: http
        - containerPort: 2379
          name: client
        - containerPort: 2380
          name: peer
        env:
        - name: ETCD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: ETCD_DATA_DIR
          value: /data/etcd
        - name: ETCD_INITIAL_CLUSTER_STATE
          value: "new"
        - name: ETCD_INITIAL_CLUSTER
          value: "go-everywhere-cluster-0=http://go-everywhere-cluster-0.go-everywhere:2380,go-everywhere-cluster-1=http://go-everywhere-cluster-1.go-everywhere:2380,go-everywhere-cluster-2=http://go-everywhere-cluster-2.go-everywhere:2380"
        volumeMounts:
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi
```

## Code Modifications Required

To support clustering, update `database.go` to read configuration from environment variables:

```go
func database() (*embed.Etcd, *clientv3.Client, *db.Client) {
    // Read from environment with defaults
    nodeName := os.Getenv("ETCD_NAME")
    if nodeName == "" {
        nodeName = "default"
    }

    dataDir := os.Getenv("ETCD_DATA_DIR")
    if dataDir == "" {
        dataDir = filepath.Join(os.TempDir(), "etcd-data")
    }

    clientURL := os.Getenv("ETCD_CLIENT_URLS")
    if clientURL == "" {
        clientURL = "http://127.0.0.1:2379"
    }

    peerURL := os.Getenv("ETCD_PEER_URLS")
    if peerURL == "" {
        peerURL = "http://127.0.0.1:2380"
    }

    initialCluster := os.Getenv("ETCD_INITIAL_CLUSTER")
    if initialCluster == "" {
        initialCluster = fmt.Sprintf("%s=%s", nodeName, peerURL)
    }

    // Configure and start embedded etcd with these values
    // ... rest of the implementation
}
```

## Load Balancing

For production deployments, place a load balancer in front of your application instances:

### HAProxy Configuration Example
```
global
    daemon

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms

frontend http_front
    bind *:80
    default_backend http_back

backend http_back
    balance roundrobin
    server node1 10.0.1.10:8000 check
    server node2 10.0.1.11:8000 check
    server node3 10.0.1.12:8000 check
```

### NGINX Configuration Example
```nginx
upstream go_everywhere_cluster {
    server 10.0.1.10:8000;
    server 10.0.1.11:8000;
    server 10.0.1.12:8000;
}

server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://go_everywhere_cluster;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Monitoring and Health Checks

### etcd Metrics
The embedded etcd exposes metrics that can be monitored:
- Endpoint: `http://<node-ip>:2379/metrics`
- Key metrics to watch:
  - `etcd_server_has_leader` - Should be 1
  - `etcd_server_leader_changes_seen_total` - Should be low
  - `etcd_network_peer_round_trip_time_seconds` - Network latency between peers

### Application Health Check
Add a health endpoint to verify cluster status:

```go
func healthCheck(client *db.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
        defer cancel()

        // Try to write and read from etcd
        testKey := "health-check"
        err := client.Put(ctx, "system", testKey, time.Now().Unix())
        if err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "unhealthy",
                "error": err.Error(),
            })
            return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
        })
    }
}
```

## Troubleshooting

### Common Issues

1. **Cluster fails to form**
   - Verify network connectivity between nodes
   - Check that all nodes have the same `ETCD_INITIAL_CLUSTER` configuration
   - Ensure ports 2379 and 2380 are not blocked by firewalls

2. **Split brain scenario**
   - Always run odd number of nodes (3, 5, 7)
   - Minimum 3 nodes for production
   - etcd requires majority (quorum) to operate

3. **Node fails to rejoin after restart**
   - If data directory exists, set `ETCD_INITIAL_CLUSTER_STATE=existing`
   - May need to remove data directory and rejoin as new member

4. **Performance issues**
   - Monitor network latency between nodes
   - Consider using SSDs for etcd data directory
   - Tune etcd parameters like heartbeat interval and election timeout

### Backup and Recovery

Regular backups are crucial for disaster recovery:

```bash
# Backup
etcdctl --endpoints=http://localhost:2379 snapshot save backup.db

# Restore (on new cluster)
etcdctl snapshot restore backup.db \
  --data-dir=/new-data-dir \
  --name=node1 \
  --initial-cluster=node1=http://localhost:2380 \
  --initial-advertise-peer-urls=http://localhost:2380
```

## Best Practices

1. **Odd number of nodes**: Always run 3, 5, or 7 nodes for proper quorum
2. **Dedicated data directory**: Use persistent storage for production
3. **Monitor cluster health**: Set up alerting for leader elections and network issues
4. **Regular backups**: Implement automated backup strategy
5. **Network reliability**: Ensure low latency (<10ms) between nodes
6. **Resource allocation**: Provide sufficient CPU and memory for etcd operations
7. **Security**: In production, use TLS for client and peer communication

## Scaling Considerations

- **Horizontal scaling**: Add more application instances as needed
- **etcd limits**: etcd performs best with 3-7 nodes
- **Data sharding**: For very large datasets, consider partitioning data across multiple etcd clusters
- **Read replicas**: Configure some nodes as read-only for better read performance

## Additional Resources

- [etcd Documentation](https://etcd.io/docs/)
- [etcd Operations Guide](https://etcd.io/docs/v3.5/op-guide/)
- [etcd Performance Tuning](https://etcd.io/docs/v3.5/tuning/)
- [etcd Disaster Recovery](https://etcd.io/docs/v3.5/op-guide/recovery/)