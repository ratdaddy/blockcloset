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
- There is no runnable code.
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

- **API / Gateway**
  Handles incoming client requests, provides the object store API, and serves the administrative UI.

- **Control Plane**
  Manages cluster membership, replication policies, and metadata about stored objects.

- **Storage Nodes**
  Store object data as blocks on local disks, participate in replication, and serve data to the gateway.

- **Management Tools**
  Expose health metrics, cluster state, and diagnostic information.

Multiple node types will exist—for example, storage nodes, gateway/control-plane nodes, or combined roles in smaller deployments.

---

## Development Notes

- **Architecture Decision Records (ADRs)**
  All significant technical decisions are documented in [`docs/adr/`](docs/adr/).
  The ADR index lists the currently active architectural decisions and their implementation status.

- **Initial Focus**
  The first milestone will be:
  1. Establishing strong foundations for development and operations, including build tooling, deployment scripts, and observability basics.
  2. Implementing the API/Gateway component with object upload/download endpoints and the beginnings of the administrative UI.
  3. Integrating basic storage-node interaction and simple replication between nodes.

---
