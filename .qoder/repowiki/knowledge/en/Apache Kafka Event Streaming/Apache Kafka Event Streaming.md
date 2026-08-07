---
kind: external_dependency
name: Apache Kafka Event Streaming
slug: apache-kafka
category: external_dependency
category_hints:
    - vendor_identity
    - framework_behavior
scope:
    - '**'
---

### Apache Kafka
- **Role**: Event streaming backbone for asynchronous data propagation between MongoDB operational writes and downstream projections.
- **Integration**: confluent-kafka-go client with segmentio/kafka-go fallback; KRaft mode enabled.
- **Usage Pattern**: Every operational CRUD write publishes events to Kafka topics, consumed by PostgreSQL builders and ClickHouse processors.
- **Architecture**: Core component of MongoDB -> Kafka -> PostgreSQL -> ClickHouse pipeline; topic names correspond to business entities (stockreceiveproduct, saleinvoice, etc.).
- **Deployment**: Confluent Local image in Docker container on port 9092.