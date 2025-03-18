# Kafka Under-Replicated Partitions
   Diagnosis:
1. Check under-replicated partitions for all topics:
    ```bash
    kafka-topics.sh --describe --bootstrap-server <brokers> --command-config config/clientauth.properties --under-replicated-partitions
    ```
2. List the broker IDs for the under-replicated partitions.
3. Verify if the broker IDs are reachable:
    ```bash
    kafka-broker-api-versions.sh --bootstrap-server <broker>
    ```
4. Check broker logs for errors.
5. Verify network connectivity between brokers.
6. Check disk space on brokers: `df -h`
7. Free up space if disks are full.

Remedy:
1. Adjust the replication factor for affected topics:
    ```bash
    kafka-topics.sh --alter --topic <topic> --replication-factor <new-value>
    ```
2. Increase the number of replicas if necessary.

# Kafka Abnormal Controller State
Diagnosis:
1. Check if only one leader controller is available:
```bash
kafka-metadata-quorum.sh --bootstrap-server <broker> --describe-controller
```
Remedy:

1. Restart the controllers.
# Kafka Offline Partitions
Diagnosis:
1. List offline partitions:
    ```bash
    kafka-topics.sh --bootstrap-server <brokers> --command-config config/clientauth.properties --describe | grep -E "Leader: -1"
    ```
2. Verify all brokers are up and reachable.
3. Check broker logs for disk or network errors.
4. Verify disk usage and health: `df -h`
5. Check network connectivity between brokers.

Remedy:
1. Restart affected brokers.
2. Investigate and resolve disk failures (e.g., replace failed drives) or network issues (e.g., check firewall rules).
3. Increase the replication factor for affected topics.

# Kafka Under Min ISR Partition Count
Diagnosis:
1. Check the `min.insync.replicas` setting for topics:
    ```bash
    kafka-configs.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties --describe --entity-type topics --all | grep min.insync.replicas
    ```
Remedy:
1. Increase min.insync.replicas for topics:
    ```bash
       kafka-configs.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties --alter --entity-type topics --entity-name <topic> --add-config min.insync.replicas=<value>
    ```
2. Optimize network performance (e.g., `reduce latency`, `increase bandwidth`).

# Kafka Offline Log Directory Count
Diagnosis:
1. Check log directory accessibility: `ls -ld /var/log/kafka/`
2. Verify disk health and space: `df -h`
3. Check log directory status:
    ```bash
    kafka-log-dirs.sh --bootstrap-server <brokers> --command-config config/clientauth.properties --describe
    ```
4. Ensure all brokers are listed and no errors are present.

Remedy:
1. Resolve disk or log directory issues (e.g., `free up space`, `repair disk errors`).
2. Check pvc, pv, and storage class for issues.

# Kafka ISR Expand Rate and Kafka ISR Shrink Rate
Diagnosis:
1. Investigate broker stability issues.
2. Check broker memory usage. Frequent broker restarts may cause ISR expansion and shrinkage.

Remedy:
1. Adjust replication settings (e.g., `replication.factor`, `min.insync.replicas`).
2. Add more brokers to balance the load and scale the cluster horizontally.
3. Find the root cause of broker instability.

# Kafka Broker Count
Diagnosis:
1. Verify all brokers are up:
    ```bash
    kafka-broker-api-versions.sh --bootstrap-server <broker-list>
    ```
2. Check for network or hardware issues.
3. 
Remedy:
1. Restart downed brokers.
2. Resolve network or hardware issues.


# Kafka Network Processor Idle Percent and Kafka Request Handler Idle Percent
Diagnosis:
1. Monitor CPU and network usage on brokers: `top`, `iftop` or any other way

Remedy:

1. Adjust thread pools in kafka reconfifure.
    - `num.network.threads`
    - `num.io.threads`
2. Optimize workload or scale brokers if overloaded.
3. Monitor NetworkProcessorIdlePercent and RequestHandlerIdlePercent.
# Kafka Replica Fetcher Manager Max Lag
Diagnosis:
1. Optimize disk I/O, network bandwidth, or fetcher threads.
2. Check network latency between brokers.

Remedy:
1. Increase fetcher threads: `num.replica.fetchers=<value>`
2. Optimize network settings (e.g., increase bandwidth, reduce latency).
3. Tune broker configuration:
4. Adjust `replica.fetch.max.bytes`
5. Adjust `replica.fetch.wait.max.ms`
6. Adjust `num.replica.fetchers` based on workload.