package server

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

func (srv *server) Withdraw(c *gin.Context) {
	// get the account by the account id from the x-api-key header
	account, err := model.GetAccountByAPIKey(c.GetHeader("x-api-key")) // you can define this in the model package
	// abort if no api key provided or account is not found
	if err != nil || account.ID == 0 {
		c.AbortWithError(401, errors.New("Unauthorized"))
		return
	}
	addressID := c.PostForm("address_id") // the address from which to withdraw the funds

	address, err := model.GetAddressByID(addressID) // @todo add method in the model
	if err != nil {
		c.AbortWithStatusJSON(404, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if address.AccountID != account.ID {
		c.AbortWithStatusJSON(404, map[string]interface{}{
			"success": false,
			"error":   "Address not found",
		})
		return
	}

	amount := c.PostForm("amount") // the amount of coins to withdraw
	symbol := c.PostForm("symbol") // the type of coin you want to withdraw
	to := c.PostForm("to")         // the address to withdraw funds to

	// create the withdraw events and send it to kafka for processing and wait for a transaction id
	addressEvent, err := srv.processWithdraw(address, amount, symbol, to)
	// send response to the user based on the address event
	if err != nil {
		c.AbortWithStatusJSON(500, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"txid":    addressEvent.TxID,
		})
		return
	}
	c.JSON(200, map[string]interface{}{
		"success": true,
		"txid":    addressEvent.TxID,
	})
}

// process withdraw request and return the created AddressEvent with the transaction id and an error
func (srv *server) processWithdraw(address *model.Address, amount, symbol, to string) (*model.AddressEvent, error) {
	var event model.AddressEvent
	event.Symbol = symbol
	event.AddressID = address.ID
	event.Type = "withdraw_request"

	// begin a transaction
	tx := s.repo.Conn.Begin()

	// check if the current balance has enough funds to execute the withdraw
	var addressBalance model.AddressBalance
	db := srv.repo.Conn.Where("address_id = ? and coin = ?", addressID, symbol).Find(&addressBalance)

	balance := s.ConvertStringToBigFloat(addressBalance.Balance)
	requestedAmount := s.ConvertStringToBigFloat(amount)

	if balance.Cmp(requestedAmount) < 0 {
		tx.Rollback()
		return event, errors.New("Insufficient funds available")
	}

	// create the withdraw_request event for the account address
	addressEvent := model.NewAddressEvent(address.ID, "withdraw_request", symbol, amount, "", "", "0", symbol)
	if err := srv.repo.Conn.Create(&addressEvent).Error; err != nil {
		tx.Rollback()
		return event, err
	}
	srv.repo.ApplyEventOnBalance(addressEvent)

	// send a withdraw request to the ethereum wallet via kafka producer
	// and wait for a response message with the transaction id
	txid, err := srv.SendWithdraw(address.PublicKey, to, symbol, amount)
	event.TxID = txid
	if err != nil {
		tx.Rollback()
		log.Println("Unable to execute withdrawal request. Failed from wallet with", err)
		return event, err
	}
	commitErr := tx.Commit().Error
	if commitErr != nil {
		log.Println("Unable to commit withdrawal request after payment was sent. Database connection issue? Transaction [", symbol, "] ", txid, commitErr)
		return event, commitErr
	}

	//in case of successful withdrawal request
	//create the withdraw event for the account address
	withdrawEvent := model.NewAddressEvent(address.ID, "withdraw", symbol, amount, txid, "", "0", symbol)
	if err := srv.repo.Conn.Create(&withdrawEvent).Error; err != nil {
		log.Println("Error adding the withdraw event for the account wallet", err)
		return event, err
	}
	srv.repo.ApplyEventOnBalance(withdrawEvent)

	return event, nil
}
