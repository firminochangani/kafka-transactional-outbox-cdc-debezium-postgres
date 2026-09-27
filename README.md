# Kafka - Transactional Outbox Pattern with CDC (Debezium and Postgres)

## Goals

- [x] Set up Kafka, Postgres and Debezium via `docker-compose`;
- [x] Write a producer that stores orders into Postgres and publishes `OrderPlaced` event into an outbox table;
- [x] Route events in the outbox table to the topic `work.firmino.orders`;
- [ ] Write a consumer that consume events from `work.firmino.orders` topic;

## Out of scope

- Routing failed messages to a DLQ;
- Hexagonal arch, DDD: This is meant to be a non-production-ready implementation;

## Setting up locally

- Clone the repository locally
- Run `make install` to install Goose to run SQL migrations;
- Run `docker compose up -d` to bring Kafka, Postgres and Debezium up;
- Run `make setup-kafka` to create the topic aforementioned;
- Run `make setup-debezium` to create the config for connector `PostgresConnector`;
- Run `make order-svc` to bring up the order service responsible for producing the event OrderPlaced;
- Run `make inventory-svc` to bring up the inventory service responsible for consuming the event OrderPlaced;
- Open Kafka [Dashboard](http://localhost:8090/topics/work.firmino.orders/?pageSize=10)
- Place an order via cURL: `TODO`;
- Run `docker compose stop` to stop the containers created by this project;