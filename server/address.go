package server

import (
	"errors"
	"time"

	"github.com/rwoj/go-vault/vault_manager/model"
	"github.com/rwoj/go-vault/vault_manager/queue"

	"github.com/gin-gonic/gin"
)

// ListAddresses list the available addresses for the user
func (srv *server) ListAddresses(c *gin.Context) {
	// get the account by the account id from the x-api-key header
	account, err := model.GetAccountByAPIKey(c.GetHeader("x-api-key")) // you can define this in the model package
	// abort if no api key provided or account is not found
	if err != nil || account.ID == 0 {
		c.AbortWithError(401, errors.New("Unauthorized"))
		return
	}
	// query the database for any address that belongs to the account
	addresss := make([]model.Address, 0, 0)
	db := srv.repo.Conn.Where("account_id = ?", account.ID).Find(&addresss)
	// abort with error if the query failed
	if db.Error != nil {
		c.AbortWithError(500, err)
		return
	}
	// return the results
	c.JSON(200, addresses)
}

// CreateAddress creates a new deposit address for the user
func (srv *server) CreateAddress(c *gin.Context) {
	// get the account by the account id from the x-api-key header
	account, err := model.GetAccountByAPIKey(c.GetHeader("x-api-key")) // you can define this in the model package
	// abort if no api key provided or account is not found
	if err != nil || account.ID == 0 {
		c.AbortWithError(401, errors.New("Unauthorized"))
		return
	}
	symbol := c.PostForm("chain_symbol")
	// send the request to the wallet through the message queue and return a reply channel where we can wait for a response
	replyChan, _ := srv.queue.ExecuteAndWait(queue.Command{
		Command:      "create_account",
		CommandTopic: symbol + ".wallet.command",
		Data: map[string]interface{}{
			"symbol": symbol,
		},
		Meta: map[string]interface{}{
			"account_id": account.ID,
		},
	})

	// wait for the response from the wallet or reply to the caller with a timeout error
	select {
	case data := <-replyChan:
		// on a success reply from the wallet save the new address in the database for the account
		publicKey, _ := data["address"].(string)
		address, err := model.CreateAddress(publicKey, symbol, account.ID) // you can define this in the model package
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		// reply with the new generated address
		c.JSON(201, address)
	case <-time.After(time.Duration(30) * time.Second):
		// timeout
		c.AbortWithError(500, errors.New("Request timeout"))
		return
	}
}

// GetAddressBalances action
func (srv *server) GetAddressBalances(c *gin.Context) {
	// get the account by the account id from the x-api-key header
	account, err := model.GetAccountByAPIKey(c.GetHeader("x-api-key")) // you can define this in the model package
	// abort if no api key provided or account is not found
	if err != nil || account.ID == 0 {
		c.AbortWithError(401, errors.New("Unauthorized"))
		return
	}
	addressID := c.PostForm("address_id")
	// load balances from the database
	addressBalances := make([]model.AddressBalance, 0, 0)
	db := srv.repo.Conn.Where("address_id = ?", addressID).Find(&addressBalances)
	if db.Error != nil {
		c.AbortWithError(500, db.Error)
		return
	}
	// display them in an easy to manage format
	balances := make(map[string]interface{}, 0)
	for _, balance := range addressBalances {
		balances[balance.Coin] = map[string]interface{}{
			"balance":       balance.Balance,
			"lockedBalance": balance.LockedBalance,
		}
	}
	c.JSON(200, balances)
}
func (srv *server) HandleDepositTransaction(symbol, txid, amount, to string) error {
	address, err := model.GetAddressByPublicKey(to) // @todo add method in the model
	if err != nil {
		return err
	}

	// add address event
	addressEvent := model.NewAddressEvent(address.ID, "deposit", symbol, amount, txid, "", "0", symbol)
	if err := srv.repo.Conn.Create(&addressEvent).Error; err != nil {
		return err
	}

	// reflect the latest event on the balance
	// this can either be done here or in a separate section in order to make the
	// event generation and the balance update completely independent
	srv.repo.ApplyEventOnBalance(addressEvent)

	return nil
}
