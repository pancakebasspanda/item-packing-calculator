# System Design Document: Item Packing Calculator

## System Overview
The Item Packing Calculator is a highly flexible Go API designed around clean architecture. Its main job is to calculate the most efficient packaging combination for a given order, ensuring that you ship the absolute lowest number of packs, send no partially filled packs, and never under-ship an order.
## Architecture Diagram

[//]: # ()
[//]: # (```mermaid)

[//]: # (flowchart TD)

[//]: # (    A["Frontend"] -->|HTTP / JSON| B["API Handler &#40;Inbound Adapter&#41;"])

[//]: # (    B --> C["Core Port Interfaces"])

[//]: # (    C --> D["Core Business Service<br/>&#40;Packing Algorithm&#41;"])

[//]: # (    C --> E["Storage Adapter<br/>&#40;In-Memory Mutex Store&#41;"])

[//]: # (    D --> F["OpenTelemetry Exporters"])

[//]: # (    F --> G["Jaeger & Prometheus"])

[//]: # ( ```)

![System Design](system-architecture.png)


## Component Architecture

| Layer / Component | Technology | Primary Responsibility |
| :--- | :--- | :--- |
| **Frontend UI** | React 18, TypeScript, Tailwind | Renders real-time packing calculations, status metrics, and quick-add presets. |
| **API Layer** | Go `net/http` | Handles input sanitization, JSON payload parsing, CORS, and rate limiting. |
| **Core Service** | Pure Go (Domain) | Executes packing optimization isolated from HTTP frameworks or storage engines. |
| **Storage Layer** | Thread-Safe Go Memory Store | Holds dynamic pack configurations (default: 250, 500, 1000, 2000, 5000) using RWMutex locks. |
| **Observability** | OpenTelemetry, Jaeger, Prometheus | Captures distributed execution spans, latency metrics, and structured JSON logs (`zerolog`). |

## Packing Algorithm Design

To balance accuracy and execution speed, the domain algorithm combines a smart recursive search with a greedy bulk-reduction strategy. For very large orders, it first uses the largest available pack sizes to fulfill the bulk of the items, which prevents memory exhaustion. It then runs the recursive search on the remaining items and seamlessly combines the two results into a single response

## API & Interface Specifications

The service follows a contract-first design driven by `openapi.yaml`.

* `POST /api/v1/packing/calculate`: Calculates optimal pack distribution for a given order quantity.
* `POST /api/v1/packs`: Adds a new pack size dynamically.
* `DELETE /api/v1/packs`: Removes a pack size dynamically.

## Resilience & Quality Assurance

* **Fuzz Testing:** Native Go fuzzing (`go test -fuzz`) tests the core domain algorithm against millions of pseudo-random inputs to prevent panic conditions and mathematical edge-case failures.
* **Trace Propagation:** OpenTelemetry context is injected at the API entry point and propagated through internal service calls to provide end-to-end trace visibility inside Jaeger.

## Future Scale Roadmap

* **Persistent Storage:** Replace the in-memory adapter with an AWS DynamoDB or PostgreSQL adapter by implementing the existing `StoragePort` interface.
* **Serverless Execution:** Compile the Go HTTP handler directly into an AWS Lambda or GCP Cloud Run deployment to reduce operational overhead and support scale-to-zero compute.

