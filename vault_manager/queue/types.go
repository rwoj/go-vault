package queue

// Queue executes commands on the wallets and can wait for replies from them
type Queue interface {
	Listen()
	Execute(command Command) error
	ExecuteAndWait(command Command) (chan map[string]interface{}, error)
	SetTransactionHandler(func(*Transaction) error)
	SetConfirmationHandler(func(*Transaction) error)
}

// Command structure
type Command struct {
	CommandTopic string                 `json:"-"`
	Command      string                 `json:"command"`
	Data         map[string]interface{} `json:"data"`
	Meta         map[string]interface{} `json:"meta"`
	ReplyTopic   string                 `json:"reply_topic,omitempty"`
}

// Transaction structure
type Transaction struct {
	Symbol string
	TxID   string
	Value  string
	To     string
}
