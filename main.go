package main

import (
	"log"
	"log/slog"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	mux := http.NewServeMux()
	port := ":8080"

	mux.Handle("/", http.HandlerFunc(homeHandler))

	slog.Info("Listening on ", "port", port)

	err := http.ListenAndServe(port, mux)
	if err != nil {
		slog.Warn("Problem starting the server", "error", err)
	}

}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	body := rabbitmqConsumer()
	w.Write([]byte(body))

}

func rabbitmqConsumer() string {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan string)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			forever <- string(d.Body)

		}
	}()

	bodyMessage := <-forever

	return bodyMessage

}
