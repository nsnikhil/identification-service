package config

type PublisherConfig struct {
	queueMap map[string]string
}

func newPublisherConfig() PublisherConfig {
	return PublisherConfig{
		queueMap: getStringMap("PUBLISHER_QUEUE_MAP"),
	}
}

func (pc PublisherConfig) QueueMap() map[string]string {
	return pc.queueMap
}
