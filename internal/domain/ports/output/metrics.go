package output

import "time"

type MetricsProvider interface {
	IncrementGRPCRequests(method, status string)
	RecordGRPCRequestDuration(method, status string, duration time.Duration)

	IncrementDatabaseQueries(queryType string, success bool)
	RecordDatabaseQueryDuration(queryType string, duration time.Duration)

	IncrementNotificationOperations(operation string, success bool)
	IncrementKafkaMessages(topic, operation string, success bool)
	RecordKafkaMessageDuration(topic, operation string, duration time.Duration)
	SetActiveConnections(count int)

	SetServiceHealth(healthy bool)
}
