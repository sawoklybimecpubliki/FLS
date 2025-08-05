package events

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"time"
)

type Event struct {
	URL string `json:"url"`
}

type ServiceNotify struct {
	BrokerAddr string
	KafkaConn  *kafka.Conn
}

const topicName = "event"
const brokerAddr = "kafka:9092" // ВРОДЕ неправильный

//var KafkaConn *kafka.Conn

func NewConnection(brokerAddr string) *kafka.Conn {
	var kafkaConn *kafka.Conn
	var err error
	for range 5 {
		kafkaConn, err = kafka.DialLeader(context.Background(), "tcp", brokerAddr, topicName, 0)
		if err == nil {
			break
		}
		log.Println("не удалось подключиться к Kafka (%s): %w", brokerAddr, err)
		time.Sleep(3 * time.Second)
	}
	log.Println("KAFKA---------------| ", kafkaConn)
	return kafkaConn
}

/*func init() {
	var once sync.Once
	log.Println("INIT: ---------------------------------------")
	once.Do(func() {
		var err error
		for range 5 {
			KafkaConn, err = kafka.DialLeader(context.Background(), "tcp", brokerAddr, topicName, 0)
			if err == nil {
				break
			}
			log.Println("не удалось подключиться к Kafka (%s): %w", brokerAddr, err)
			time.Sleep(3 * time.Second)
		}
	})
	log.Println("KAFKA---------------| ", KafkaConn)
}*/

func (s *ServiceNotify) ProduceEvent(event *Event) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	msg := kafka.Message{
		Value: eventJSON,
		Time:  time.Now(),
	}

	_, err = s.KafkaConn.WriteMessages(msg)
	if err != nil {
		return fmt.Errorf("ошибка записи сообщения в Kafka: %w", err)
	}
	return nil
}

func (s *ServiceNotify) ReceiverEvent(ctx context.Context) []string {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddr},
		Topic:    topicName,
		MinBytes: 10e3,
		MaxBytes: 10e6,
		MaxWait:  1 * time.Second,
	})

	var out []string
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		out = append(out, string(m.Value))
		fmt.Printf("Message received from topic %s, partition %d, offset %d: %s\n",
			m.Topic, m.Partition, m.Offset, string(m.Value))

		if err := r.Close(); err != nil {
			log.Fatal("failed to close reader:", err)
			break
		}
	}
	return out
}

func (s *ServiceNotify) EventsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.ProduceEvent(&Event{
			URL: r.URL.Path,
		}); err != nil {
			log.Println("error send: ", err)
		}

		next.ServeHTTP(w, r)
	}
}
