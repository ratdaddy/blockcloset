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
curl -i -X PUT http://$FLATBED_ADDR/hello

# create bucket with an invalid name:
curl -i -X PUT http://$FLATBED_ADDR/invalid_name!

# create bucket with an invalid bucket name (as seen by gantry):
curl -i -X PUT http://$FLATBED_ADDR/bad

# create bucket with a bucket name we don't own:
curl -i -X PUT http://$FLATBED_ADDR/taken

# panic gantry:
curl -i -X PUT http://$FLATBED_ADDR/panic

# list buckets:
curl -i http://$FLATBED_ADDR/

# put object:
curl -i -X PUT --data 'hello' http://$FLATBED_ADDR/hello/object

# put object from a file:
curl -X PUT -H "Content-Length: $(wc -c < filename)" --data-binary @filename http://$FLATBED_ADDR/hello/object

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

# put object with a bucket name we don't own:
curl -i -X PUT --data 'hello' http://$FLATBED_ADDR/forbidden/object

# put object with an invalid key name:
curl -i -X PUT -d 'hello' http://$FLATBED_ADDR/my-bucket/$(head -c 1025 /dev/zero | tr '\0' 'a')

# put object with no cradle servers:
curl -i -X PUT --data 'hello' http://$FLATBED_ADDR/no-cradles/object

# put object panic gantry:
curl -i -X PUT --data 'hello' http://$FLATBED_ADDR/panic/object

# put object with write size mismatch
curl -i -X PUT --data 'hello' http://$FLATBED_ADDR/size-mismatch/object
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

# resolve write:
grpcurl -plaintext -d '{"bucket":"my-bucket","key":"my-key.txt","size":1024}' $GANTRY_ADDR gantry.service.v1.GantryService/PlanWrite
```

Grpcurl exampe to run directly with cradle:
```bash
# reflection:
grpcurl -plaintext $CRADLE_ADDR list cradle.service.v1.CradleService
grpcurl -plaintext $CRADLE_ADDR describe cradle.service.v1.CradleService.WriteObject

# successful write object
cat <<'EOF' | \
grpcurl -plaintext -d @ $CRADLE_ADDR cradle.service.v1.CradleService/WriteObject
{"metadata":{"object_id":"test123","bucket":"test-bucket","size":11}}
{"chunk":"aGVsbG8g"}
{"chunk":"d29ybGQ="}
EOF

# stream failure
grpcurl -plaintext -max-time 15 -d @ $CRADLE_ADDR cradle.service.v1.CradleService/WriteObject
# Then let it time out or type and wait for timeout:
{"metadata": {"object_id": "test-123", "bucket": "photos", "size": 11}}

# metadata not first
echo '{"chunk":"aGVsbG8g"}' | \
grpcurl -plaintext -d @ $CRADLE_ADDR cradle.service.v1.CradleService/WriteObject
```
