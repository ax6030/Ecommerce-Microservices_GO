package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchange    = "order.events"
	dlxExchange = "order.events.dlx"
	queue       = "notification.order.created"
	dlqQueue    = "notification.order.created.dlq"
)

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewConsumer(url string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err := setupTopology(ch); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &Consumer{conn: conn, ch: ch}, nil
}

func setupTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(dlxExchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(dlqQueue, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(dlqQueue, "order.created.dead", dlxExchange, false, nil); err != nil {
		return err
	}

	if err := ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	args := amqp.Table{
		"x-dead-letter-exchange":     dlxExchange,
		"x-dead-letter-routing-key":  "order.created.dead",
	}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, args); err != nil {
		return err
	}
	return ch.QueueBind(queue, "order.created", exchange, false, nil)
}

func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	return c.ch.Consume(queue, "", false, false, false, false, nil)
}

func (c *Consumer) Close() {
	c.ch.Close()
	c.conn.Close()
}
