package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ahugofreire/car-simulator-freight/common"
	"github.com/ahugofreire/car-simulator-freight/internal/freight/entity"
	"github.com/ahugofreire/car-simulator-freight/internal/freight/infra/repository"
	"github.com/ahugofreire/car-simulator-freight/internal/freight/usecase"
	"github.com/ahugofreire/car-simulator-freight/pkg/kafka"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", common.MysqlDNS)
	if err != nil {
		panic(err)
	}
	defer db.Close()

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
			output, err := createRouteUseCase.Execute(input)
			if err != nil {
				log.Println(err)
			}
			fmt.Println(output)

		case "RouteStarted", "RouteFinished":
			fmt.Println(input)
			input := usecase.ChangeRouteStatusInput{}
			json.Unmarshal(message.Value, &input)
			output, err := changeRouteStatusUseCase.Execute(input)
			if err != nil {
				log.Println(err)
			}
			fmt.Println(output)
		}
	}
}
