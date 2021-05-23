package server

import (
	cgf "vault_manager/config"
	"vault_manager/model"
	"vault_manager/net"
	"vault_manager/queue"

	"github.com/gin-gonic/gin"
)

// Server interface
type Server interface {
	Listen()
}

// server structure
type server struct {
	config    cgf.Config
	queue     queue.Queue
	repo      *model.Repo
	consumers map[string]net.KafkaConsumer
	producers map[string]net.KafkaProducer
}

// NewServer creates a new server instance that can process the requests
func NewServer(config cgf.Config) Server {
	// init producers
	producers := make(map[string]net.KafkaProducer, len(config.Brokers.Producers))
	for key, brokerCfg := range config.Brokers.Producers {
		producers[key] = NewProducer(brokerCfg)
	}

	// init consumers
	consumers := make(map[string]net.KafkaConsumer, len(config.Brokers.Consumers))
	for key, brokerCfg := range config.Brokers.Consumers {
		consumers[key] = NewConsumer(brokerCfg)
	}

	// setup database connection
	dbConf := config.Database
	repo := model.NewRepo(dbConf.Host, dbConf.Port, dbConf.Username, dbConf.Password, dbConf.Name)

	// connect the queue to the right producer and consumer
	q := queue.NewQueue(producers["commands"], consumers["events"], config.Chains)

	// return the server structure
	return &server{
		config:    config,
		queue:     q,
		repo:      repo,
		producers: producers,
		consumers: consumers,
	}
}

// Listen to requests from the API
func (srv *server) ListenToRequests() {
	r := gin.Default()
	srv.queue.SetTransactionHandler(func(transaction *queue.Transaction) error {
		return srv.HandleDepositTransaction(transaction.Symbol, transaction.TxID, transaction.Value, transaction.To)
	})

	// I should be able to add a new address for an existing chain
	r.POST("/address", srv.CreateAddress)
	r.GET("/address", srv.ListAddresses)
	r.GET("/address/:address_id/balance", srv.GetAddressBalances)

	// I should be able to withdraw funds from a generated address
	r.POST("/withdraw/address/:address_id", srv.Withdraw)

	r.Run() // listen and serve on 0.0.0.0:80
}
