package nats

import nc "github.com/nats-io/nats.go"

func CreateConsumer(js nc.JetStreamContext, topic string, durableName string) (*nc.ConsumerInfo, error) {
	return js.AddConsumer(topic, &nc.ConsumerConfig{
		Durable:        durableName,
		DeliverPolicy:  nc.DeliverAllPolicy,
		AckPolicy:      nc.AckExplicitPolicy,
		FilterSubject:  topic,
		ReplayPolicy:   nc.ReplayInstantPolicy,
		DeliverSubject: nc.NewInbox(),
		DeliverGroup:   durableName + "-" + topic,
	})
}
