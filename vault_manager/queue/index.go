// define the queue structure
type queue struct {
	config             map[string]config.ChainConfig
	producer           net.KafkaProducer
	consumer           net.KafkaConsumer
	pending            map[string]chan map[string]interface{}
	transactionHandler func(*Transaction) error
	confirmTransaction func(*Transaction) error
}

// NewQueue constructor
func NewQueue(producer net.KafkaProducer, consumer net.KafkaConsumer, cfg map[string]config.ChainConfig) Queue {
	return &queue{
		producer: producer,
		consumer: consumer,
		config:   cfg,
		pending:  make(map[string]chan map[string]interface{}),
	}
}

// set a handler for any incomming transaction trasaction
func (q *queue) SetTransactionHandler(transactionHandler func(*Transaction) error) {
	q.transactionHandler = transactionHandler
}

// if the message comes from a ""<chain_symbol>.transaction" topic then
// create a new transaction object from it and pass it to the handler
func (q *queue) ProcessTransaction(symbol string, msg *sarama.ConsumerMessage) {
	data := make(map[string]interface{})
	err := json.Unmarshal(msg.Value, &data)
	if err != nil {
		log.Println("Error parsing transaction", err, msg)
		return
	}

	transaction := &Transaction{
		Symbol: data["symbol"].(string),
		TxID:   data["txid"].(string),
		Value:  data["value"].(string),
		To:     data["to"].(string),
	}
	go q.transactionHandler(transaction)
}

// other methods from the interface
// Listen()
// Execute(command Command) error
// ExecuteAndWait(command Command) (chan map[string]interface{}, error)
// ...