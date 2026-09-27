# Kafka - Transactional Outbox Pattern with CDC (Debezium and Postgres)

## Goals

- Set up Kafka, Postgres and Debezium via `docker-compose`;
- Write a producer that stores orders into Postgres and publishes `OrderPlaced` event into an outbox table;
- Route events in the outbox table to the topic `work.firmino.orders`;
- Write a consumer that consume events from `work.firmino.orders` topic;

## Out of scope

- Routing failed DLQ