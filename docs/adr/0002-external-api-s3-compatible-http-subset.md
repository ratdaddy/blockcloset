# ADR 0002: External API – S3-Compatible HTTP Subset

* **Status**: Adopted
* **Date**: 2025-08-10
* **Author**: Brian VanLoo

## Context

Amazon S3 provides a proven, widely adopted API and a valuable service model. Building Block Closet on a subset of the S3 API leverages this maturity while allowing selective modernization where appropriate. S3 defines a REST-style HTTP interface for bucket operations, object operations, and multipart uploads, with authentication via AWS Signature Version 4 (SigV4). Block Closet will begin by implementing these core definitions.

Block Closet will use path-based routing for initial compatibility and simplicity, as virtual-hosted–style URLs are more complex to implement in a homelab without integrated DNS services. While S3 is moving away from path-based routing, Block Closet will adopt it initially for practical reasons.

## Decision

Block Closet will implement an **S3-compatible HTTP subset** as the primary external API. The initial focus will be on:

1. **Bucket Operations**

   * Create/Delete bucket
   * List buckets
   * List objects in a bucket (ListObjectsV2)

2. **Object Operations**

   * PUT/GET/HEAD/DELETE Object
   * Basic metadata support (Content-Length, Content-Type, ETag)

3. **Multipart Upload**

   * Initiate (CreateMultipartUpload)
   * UploadPart (and UploadPartCopy in a later phase)
   * CompleteMultipartUpload
   * AbortMultipartUpload

4. **Authentication**

   * Implement AWS SigV4 signing for future compatibility, but begin with all access anonymous until authentication is enabled.

5. **Request Addressing**

   * Use path-style addressing for all requests to simplify initial homelab deployment.

### Wire Format & Modernization Approach

* **Compatibility First**: For endpoints where S3 clients expect XML payloads (e.g., ListBuckets, ListObjectsV2, CompleteMultipartUpload), Block Closet will return XML matching S3 schemas to ensure tool compatibility.
* **JSON for Block Closet-Native Endpoints**: The admin UI and any Block Closet-specific APIs outside the S3 subset will use JSON. Optional JSON responses for certain S3-compatible endpoints may be offered later via content negotiation without breaking existing S3 clients.

## Consequences

* **Proven Foundation**: Builds on a well-documented and recognized API.
* **Tool Compatibility**: Enables AWS CLI/SDKs and many third-party clients to work with Block Closet for core operations.
* **Simplified Deployment**: Path-style addressing eliminates DNS requirements for initial homelab use.
* **Security Posture**: Starts with anonymous access for ease of setup; SigV4 authentication will be available when enabled.
* **Implementation Complexity**: XML responses add effort but preserve interoperability.

## Alternatives Considered

* **Custom HTTP+JSON API Only**

  * Rejected for now; easier to implement but forfeits compatibility with the S3 ecosystem.

* **gRPC as the Primary API**

  * Rejected initially; suitable for service-to-service communication but not for existing S3 clients.

* **JSON-Only S3 Compatibility**

  * Rejected; many tools depend on XML responses for core operations.

* **Virtual-Hosted Style URLs**

  * Rejected initially; introduces DNS complexity not suited to early homelab deployments.

## References

* [S3 REST API – Making Requests](https://docs.aws.amazon.com/AmazonS3/latest/API/RESTAPI.html)
* [S3 API Reference – Bucket/Object Operations](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Operations_Amazon_Simple_Storage_Service.html)
* [Multipart Upload Overview & Operations](https://docs.aws.amazon.com/AmazonS3/latest/userguide/mpuoverview.html)
* [SigV4 Authentication (S3)](https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html)
* [Virtual-Hosted vs Path-Style Addressing](https://docs.aws.amazon.com/AmazonS3/latest/userguide/VirtualHosting.html)

