# Server Command 

## Kafka Command


**List Topic**
```
./kafka-topics.sh --bootstrap-server localhost:9092 --list
```

**Consume Topic**
```
./kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic when-inventory-created
```

**Produce Message**
```
./kafka-console-producer.sh --bootstrap-server localhost:9092 --topic when-inventory-created
```
