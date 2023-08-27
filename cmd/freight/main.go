package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ahugofreire/car-simulator-freight/common"
	"github.com/ahugofreire/car-simulator-freight/internal/freight/entity"
	"github.com/ahugofreire/car-simulator-freight/internal/freight/infra/repository"
	"github.com/ahugofreire/car-simulator-freight/internal/freight/usecase"
	"github.com/ahugofreire/car-simulator-freight/pkg/kafka"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	_ "github.com/go-sql-driver/mysql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	routesCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "routes_created_total",
			Help: "Total number of created routes",
		},
	)

	routesStarted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "routes_started_total",
			Help: "Total number of started routes",
		},
		[]string{"status"},
	)

	errorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
	)

	routesFinished = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "routes_finished_total",
			Help: "Total number of finished routes",
		},
	)
)

func init() {
	prometheus.MustRegister(routesCreated)
	prometheus.MustRegister(routesStarted)
	prometheus.MustRegister(routesFinished)
	prometheus.MustRegister(errorsTotal)
}

func main() {
	db, err := sql.Open("mysql", common.MysqlDNS)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		fmt.Println("Server running in port 8080!!")
		http.ListenAndServe(":8080", nil)
	}()

	msgChan := make(chan *ckafka.Message)
	topics := []string{"route"}
	servers := common.KafkaServer

	go kafka.Consume(topics, servers, msgChan)

	repository := repository.NewRouteRepositoryMysql(db)
	freight := entity.NewFreight(10)
	createRouteUseCase := usecase.NewCreateRouteUseCase(repository, freight)
	changeRouteStatusUseCase := usecase.NewChangeRouteUseCase(repository)

	for message := range msgChan {
		input := usecase.CreateRouteInput{}
		json.Unmarshal(message.Value, &input)

		switch input.Event {
		case "RouteCreated":
			log.Printf("Received route created [%v]", input.ID)
			output, err := createRouteUseCase.Execute(input)
			if err != nil {
				log.Println(err)
			} else {
				fmt.Println(output)
				routesCreated.Inc()
			}
			fmt.Println(output)

		case "RouteStarted":
			log.Printf("Received route started [%v]", input.ID)
			input := usecase.ChangeRouteStatusInput{}
			json.Unmarshal(message.Value, &input)
			output, err := changeRouteStatusUseCase.Execute(input)
			if err != nil {
				log.Println(err)
				errorsTotal.Inc()
			} else {
				routesStarted.WithLabelValues("started").Inc()
				log.Println(output)
			}

		case "RouteFinished":
			log.Printf("Received route finished [%v]", input.ID)

			input := usecase.ChangeRouteStatusInput{}
			json.Unmarshal(message.Value, &input)
			_, err := changeRouteStatusUseCase.Execute(input)
			if err != nil {
				log.Println(err)
				errorsTotal.Inc()
			}

			routesFinished.Inc()
		}
	}
}
