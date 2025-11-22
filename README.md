# Block Closet

Block Closet is an object storage system designed to make use of spare disk space in a homelab.
It runs as a cluster of storage agents on multiple hosts, pooling available space into a unified storage service.
Data is replicated across the cluster to ensure high availability and durability.

The project is intended to be **easy to configure and manage**, with:
- A web-based UI for setup, management, and monitoring.
- Built-in tools for assessing cluster health.
- Diagnostics for debugging anomalous situations and performance issues.

---

## Setup

Shared environment configuration is managed via [direnv](https://direnv.net/).
The environment variables themselves are defined in `env/.env.<environment>` so for development it's `env/.env.development`.
Local overrides can be placed in `env/.env.local`.

---

## Current Status

This repository is in the earliest stage of development.
At present:
- Initial documentation and architectural groundwork are being established.
- The first ADRs are being written to guide implementation.

---

## Planned Capabilities

- **Object Storage**
  Store and retrieve objects via a simple API.

- **Distributed Operation**
  Multiple agents running on different hosts cooperate to present a single logical storage system.

- **Space Reclamation**
  Make use of unused disk capacity on each host without requiring dedicated storage hardware.

- **Data Replication**
  Objects are replicated across nodes to protect against data loss from a single host failure.

- **Cluster Management**
  Automatic node discovery, health monitoring, and rebalancing of data as storage availability changes.

---

## High-Level Architecture

The system is composed of several cooperating components:

- **API: Flatbed** -
  Handles incoming client requests, provides the object store API, and serves the administrative UI.

- **Control Plane: Gantry** -
  Manages cluster membership, replication policies, and metadata about stored objects.

- **Storage Nodes: Cradle** -
  Store object data as blocks on local disks, participate in replication, and serve data to the flatbed.

- **Management Tools** -
  Expose health metrics, cluster state, and diagnostic information.

Multiple node types will exist—for example, storage nodes, flatbed/control-plane nodes, or combined roles in smaller deployments.

---

## Development Notes

- **Architecture Decision Records (ADRs)**
  All significant technical decisions are documented in [`docs/adr/`](docs/adr/).
  The ADR index lists the currently active architectural decisions and their implementation status.

- **Initial Focus**
  The first milestone will be:
  1. Establishing strong foundations for development and operations, including build tooling, deployment scripts, and observability basics.
  2. Implementing the API/Flatbed component with object upload/download endpoints and the beginnings of the administrative UI.
  3. Integrating basic storage-node interaction and simple replication between nodes.

---

## Running locally

In the flatbed and gantry directories run the app with `make run`.

Curl examples to run through flatbed:
```bash
# create bucket:
curl -i -X PUT --data '' http://$FLATBED_ADDR/hello

# create bucket with an invalid name:
curl -i -X PUT --data '' http://$FLATBED_ADDR/invalid_name!

# create bucket with an invalid filename (as seen by gantry):
curl -i -X PUT --data '' http://$FLATBED_ADDR/bad

# panic gantry:
curl -i -X PUT --data '' http://$FLATBED_ADDR/panic

# list buckets:
curl -i http://$FLATBED_ADDR/

# put object:
curl -i -X PUT --data 'hello' http://$FLATBED_ADDR/hello/object

# put object with no content length error:
curl -i -X PUT --data '' http://$FLATBED_ADDR/hello/object -H "Content-Length:" --http1.0

# put object with 0-length content:
curl -i -X PUT --data '' http://$FLATBED_ADDR/hello/object

# put object with content length too large error:
curl -i -X PUT --data '' http://$FLATBED_ADDR/hello/object -H "Content-Length: 5368709121"

# put object with invalid transfer encoding:
  curl -i -X PUT -H "Content-Length: 1024" -H "Transfer-Encoding: chunked" http://$FLATBED_ADDR/hello/object

# put object with an invalid bucket name:
curl -i -X PUT --data 'hello' http://$FLATBED_ADDR/Invalid-name/object

# put object with an invalid key name:
curl -i -X PUT -d 'hello' http://$FLATBED_ADDR/my-bucket/$(head -c 1025 /dev/zero | tr '\0' 'a')
```

Grpcurl example to run directly with gantry:
```bash
# reflection:
grpcurl -plaintext $GANTRY_ADDR list gantry.service.v1.GantryService
grpcurl -plaintext $GANTRY_ADDR describe gantry.service.v1.GantryService.ListBuckets

# successful bucket creation:
grpcurl -plaintext -d '{"name":"bucket"}' $GANTRY_ADDR gantry.service.v1.GantryService/CreateBucket

# attempt to create a bucket with an invalid name:
grpcurl -plaintext -d '{"name":"bad"}' $GANTRY_ADDR gantry.service.v1.GantryService/CreateBucket

# panic:
grpcurl -plaintext -d '{"name":"panic"}' $GANTRY_ADDR gantry.service.v1.GantryService/CreateBucket

# list buckets:
grpcurl -plaintext -d '{}' $GANTRY_ADDR gantry.service.v1.GantryService/ListBuckets
---
