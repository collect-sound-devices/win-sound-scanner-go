package scannerapp

const (
	EnvWinSoundEnqueuer              = "WIN_SOUND_ENQUEUER"
	EnvWinSoundEnqueuerVal00Empty    = "empty"
	EnvWinSoundEnqueuerVal01RabbitMq = "rabbitmq"
	EnvWinSoundEnqueuerVal02Kafka    = "kafka"
	EnvWinSoundRabbitMQHost          = "WIN_SOUND_RABBITMQ_HOST"
	EnvWinSoundRabbitMQPort          = "WIN_SOUND_RABBITMQ_PORT"
	EnvWinSoundRabbitMQVHost         = "WIN_SOUND_RABBITMQ_VHOST"
	EnvWinSoundRabbitMQUser          = "WIN_SOUND_RABBITMQ_USER"
	EnvWinSoundRabbitMQPassword      = "WIN_SOUND_RABBITMQ_PASSWORD"
	EnvWinSoundRabbitMQExchange      = "WIN_SOUND_RABBITMQ_EXCHANGE"
	EnvWinSoundRabbitMQQueue         = "WIN_SOUND_RABBITMQ_QUEUE"
	EnvWinSoundRabbitMQRoutingKey    = "WIN_SOUND_RABBITMQ_ROUTING_KEY"
	EnvWinSoundKafkaBrokers          = "WIN_SOUND_KAFKA_BROKERS"
	EnvWinSoundKafkaTopic            = "WIN_SOUND_KAFKA_TOPIC"
	EnvWinSoundKafkaClientID         = "WIN_SOUND_KAFKA_CLIENT_ID"
	EnvWinSoundKafkaWriteTimeout     = "WIN_SOUND_KAFKA_WRITE_TIMEOUT_MS"
)
