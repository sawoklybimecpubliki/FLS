package kafka_events

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"sync"
	"time"
)

type Event struct {
	URL string `json:"url"`
}

const topicName = "event"
const brokerAddr = "kafka:9092" // ВРОДЕ неправильный

var kafkaConn *kafka.Conn

func init() {
	var once sync.Once
	log.Println("INIT: ---------------------------------------")
	once.Do(func() {
		var err error
		kafkaConn, err = kafka.DialLeader(context.Background(), "tcp", brokerAddr, topicName, 0)
		if err != nil {
			log.Println("не удалось подключиться к Kafka (%s): %w", brokerAddr, err)
		}
	})
	log.Println("KAFKA---------------| ", kafkaConn)
}

func ProduceEvent(conn *kafka.Conn, event *Event) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	msg := kafka.Message{
		Value: eventJSON,
		Time:  time.Now(),
	}

	_, err = conn.WriteMessages(msg)
	if err != nil {
		return fmt.Errorf("ошибка записи сообщения в Kafka: %w", err)
	}
	return nil
}

func EventsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ProduceEvent(kafkaConn, &Event{
			URL: r.URL.Path,
		}); err != nil {
			log.Println("error send: ", err)
		}

		next.ServeHTTP(w, r)
	}
}
