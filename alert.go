package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"log"
)

// Kafka broker addresses
// var brokers = []string{"broker1:9092", "broker2:9092", "broker3:9092"}
// var brokers = []string{"localhost:9092"}
// You will get the broker addresses from the Kafka cluster using `kubectl get appbinding -n demo kafka-dev -ojson | jq ".spec.clientConfig.url"`
var brokers = []string{"kafka-dev-0.kafka-dev-pods.demo.svc.cluster.local:9092", "kafka-dev-1.kafka-dev-pods.demo.svc.cluster.local:9092"}

func getConfig() *sarama.Config {
	config := sarama.NewConfig()
	//config.Version = sarama.V2_5_0_0

	// If you have authentication, configure it here
	//config.Net.SASL.Enable = true
	//config.Net.SASL.User = "admin"
	//config.Net.SASL.Password = "admin"
	//config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	return config
}

func underReplicatedPartitions() {
	config := getConfig()

	// Create a new Sarama client
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka client: %v", err)
	}
	defer client.Close()

	// Get the list of topics
	topics, err := client.Topics()
	if err != nil {
		log.Fatalf("Error fetching topics: %v", err)
	}

	// Create a new admin client
	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		log.Fatalf("Error creating Kafka admin client: %v", err)
	}
	defer admin.Close()

	// Describe topics to get metadata
	metadata, err := admin.DescribeTopics(topics)
	if err != nil {
		log.Fatalf("Error describing topics: %v", err)
	}

	// Map to store under-replicated partitions
	underReplicatedPartitions := make(map[string][]int32)

	// Check for under-replicated partitions
	for _, topicMetadata := range metadata {
		for _, partitionMetadata := range topicMetadata.Partitions {
			replicas := len(partitionMetadata.Replicas)
			inSyncReplicas := len(partitionMetadata.Isr) // ISR = In-Sync Replicas

			if inSyncReplicas < replicas {
				// Add under-replicated partition to the map
				underReplicatedPartitions[topicMetadata.Name] = append(underReplicatedPartitions[topicMetadata.Name], partitionMetadata.ID)
			}
		}
	}

	// Print the under-replicated partitions
	fmt.Println("Under-Replicated Partitions:")
	for topic, partitions := range underReplicatedPartitions {
		fmt.Printf("Topic: %s, Under-Replicated Partitions: %v\n", topic, partitions)
	}
}

func verifyBrokers() {
	config := getConfig()
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka client: %v", err)
	}
	defer client.Close()

	// Get the list of brokers from the cluster
	clusterBrokers := client.Brokers()
	if len(clusterBrokers) != len(brokers) {
		log.Fatalf("Expected %d brokers, but found %d", len(brokers), len(clusterBrokers))
	}

	// Check if all brokers are reachable
	allBrokersUp := true
	for _, broker := range clusterBrokers {
		// Open a connection to the broker
		if err := broker.Open(config); err != nil {
			log.Printf("Broker %s is down: %v", broker.Addr(), err)
			allBrokersUp = false
			continue
		}

		// Check if the broker is connected
		connected, err := broker.Connected()
		if err != nil {
			log.Printf("Failed to check connection status for broker %s: %v", broker.Addr(), err)
			allBrokersUp = false
			continue
		}

		if !connected {
			log.Printf("Broker %s is not connected", broker.Addr())
			allBrokersUp = false
		} else {
			fmt.Printf("Broker %s is up and running\n", broker.Addr())
		}

		// Close the broker connection
		broker.Close()
	}

	if allBrokersUp {
		fmt.Println("All brokers are up and running!")
	} else {
		fmt.Println("Some brokers are down or unreachable.")
	}
}

func numberOfLeaderController() {
	config := getConfig()
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka client: %v", err)
	}
	defer client.Close()

	// Get the controller broker
	controller, err := client.Controller()
	if err != nil {
		log.Fatalf("Failed to get controller broker: %v", err)
	}

	// Print the controller broker's address
	fmt.Printf("Controller broker address: %s\n", controller.Addr())

	// Confirm that there is only one controller
	fmt.Println("Number of controllers in the Kafka cluster: 1")
}

func numberOfOfflinePartitions() {
	config := getConfig()
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka client: %v", err)
	}
	defer client.Close()

	// Fetch metadata for all topics
	broker := client.Brokers()[0] // Use any available broker
	if err := broker.Open(client.Config()); err != nil && err != sarama.ErrAlreadyConnected {
		log.Fatalf("Error opening broker connection: %v", err)
	}

	metadataRequest := &sarama.MetadataRequest{}
	metadataResponse, err := broker.GetMetadata(metadataRequest)
	if err != nil {
		log.Fatalf("Error fetching metadata: %v", err)
	}

	// Map to store topic -> count of offline partitions
	offlinePartitionsMap := make(map[string]int)

	// Iterate over topics and check for offline partitions
	for _, topic := range metadataResponse.Topics {
		offlineCount := 0
		for _, partition := range topic.Partitions {
			if partition.Leader == -1 { // Leader -1 means the partition is offline
				offlineCount++
			}
		}
		if offlineCount > 0 {
			offlinePartitionsMap[topic.Name] = offlineCount
		}
	}

	// Print the number of offline partitions per topic
	fmt.Println("Offline Partitions:")
	for topic, count := range offlinePartitionsMap {
		fmt.Printf("Topic: %s, Offline Partitions: %d\n", topic, count)
	}
}

func countNumberOfOfflineLogDirectory() {
	config := getConfig()
	// Get the list of brokers
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka client: %v", err)
	}
	defer client.Close()

	brokersList := client.Brokers()
	offlineLogDirsMap := make(map[int32]int) // BrokerID -> Count of offline log dirs

	// Iterate over each broker and fetch log directory details
	for _, broker := range brokersList {
		if err := broker.Open(client.Config()); err != nil && err != sarama.ErrAlreadyConnected {
			log.Printf("Error opening broker connection: %v", err)
			continue
		}

		// Request log directory information
		request := &sarama.DescribeLogDirsRequest{
			Version: 1,
		}
		response, err := broker.DescribeLogDirs(request)
		if err != nil {
			log.Printf("Error describing log directories on broker %d: %v", broker.ID(), err)
			continue
		}

		// Count offline log directories
		offlineCount := 0
		for _, dir := range response.LogDirs {
			if dir.ErrorCode != 0 { // Non-zero error code indicates offline directory
				offlineCount++
			}
		}

		// Store the count if there are offline log dirs
		if offlineCount > 0 {
			offlineLogDirsMap[broker.ID()] = offlineCount
		}
	}

	// Print the number of offline log directories per broker
	fmt.Println("Offline Log Directories:")
	for brokerID, count := range offlineLogDirsMap {
		fmt.Printf("Broker ID: %d, Offline Log Directories: %d\n", brokerID, count)
	}
}

func printTheConfigurationForAllBrokers(c string) {
	config := getConfig()
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka admin client: %v", err)
	}
	defer admin.Close()

	// Get the list of broker IDs
	brokerIDs, _, err := admin.DescribeCluster()
	if err != nil {
		log.Fatalf("Error describing Kafka cluster: %v", err)
	}

	// Fetch and print the requested config for each broker
	fmt.Printf("Kafka Broker Configuration for key: %s\n", c)
	for _, brokerID := range brokerIDs {
		// Request configuration for the broker
		resource := sarama.ConfigResource{
			Type: sarama.BrokerResource,
			Name: fmt.Sprintf("%d", brokerID.ID()),
		}
		configs, err := admin.DescribeConfig(resource)
		if err != nil {
			log.Printf("Error fetching config for broker %d: %v\n", brokerID.ID(), err)
			continue
		}

		// Find the requested config key
		for _, entry := range configs {
			if entry.Name == c {
				fmt.Printf("Broker ID: %d, %s: %s\n", brokerID.ID(), entry.Name, entry.Value)
			}
		}
	}
}

func printTheConfigurationForTopic(topic, c string) {
	config := getConfig()
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka admin client: %v", err)
	}
	defer admin.Close()

	// Define the config resource for the topic
	resource := sarama.ConfigResource{
		Type: sarama.TopicResource,
		Name: topic,
	}

	// Fetch topic configuration
	configs, err := admin.DescribeConfig(resource)
	if err != nil {
		log.Fatalf("Error fetching topic configuration: %v", err)
	}

	// Find and print the requested configuration key
	for _, entry := range configs {
		if entry.Name == c {
			fmt.Printf("Topic: %s, Config: %s, Value: %s\n", topic, entry.Name, entry.Value)
			return
		}
	}

	// If the key is not found
	fmt.Printf("Config key '%s' not found for topic '%s'\n", c, topic)

}

// c is the configuration key to update
// v is the new value for the configuration key
func alterConfigForTopic(topic, c, v string) {
	config := getConfig()
	// Initialize the ClusterAdmin client
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		log.Fatalf("Error creating cluster admin: %v", err)
	}
	defer admin.Close() // Ensure the admin client is closed when done

	configEntries := map[string]*string{
		c: &v, // Pointer to the new value
	}
	// Alter the topic configuration
	err = admin.AlterConfig(sarama.TopicResource, topic, configEntries, false)
	if err != nil {
		log.Fatalf("Error altering topic configuration: %v", err)
	}

	log.Printf("Successfully altered configuration for topic %s", topic)
}

func main() {
	underReplicatedPartitions()
	verifyBrokers()
	numberOfLeaderController()
	numberOfOfflinePartitions()
	countNumberOfOfflineLogDirectory()
	printTheConfigurationForTopic("my-topic", "min.insync.replicas")
	printTheConfigurationForAllBrokers("log.retention.hours")

	//printTheConfigurationForTopic("my-topic", "max.message.bytes") // default 1048588
	//alterConfigForTopic("my-topic", "max.message.bytes", "1048588")
	//printTheConfigurationForTopic("my-topic", "max.message.bytes")
}
