import (
	"math/big"
	"vault_manager/model"
)

func (repo *Repo) ApplyEventOnBalance(addressEvent *model.AddressEvent) error {
	// check if an address balance exists for the symbol and add it if it does not
	addressBalance := model.AddressBalance{}
	exists := tx.Where("address_id = ? and coin = ?", addressEvent.AddressID, addressEvent.Symbol).Find(&addressBalance).RecordNotFound()
	if !exists {
		addressBalance := model.NewAddressBalance(symbol, address.ID, addressEvent.ID, "0", "0")
		err := repo.Conn.Create(&addressBalance).Error
		if err != nil {
			return err
		}
	}

	addressBalance.LastEventID = addressEvent.ID

	switch addressEvent.Type {
	case "deposit":
		// apply the deposit event on the address balance
		balance := ConvertStringToBigFloat(addressBalance.Balance)
		amount := ConvertStringToBigFloat(addressEvent.Amount)
		balance.Add(balance, amount)

		addressBalance.Balance = balance.String()
		return repo.Conn.Save(&addressBalance).Error

	case "withdraw_request":
		// apply a withdraw request on the balance and lock the funds
		lockedBalance := ConvertStringToBigFloat(addressBalance.LockedBalance)
		balance := ConvertStringToBigFloat(addressBalance.Balance)
		amount := ConvertStringToBigFloat(addressEvent.Amount)
		lockedBalance.Add(lockedBalance, amount)
		balance.Sub(balance, amount)

		addressBalance.LockedBalance = lockedBalance.String()
		addressBalance.Balance = balance.String()
		return repo.Conn.Save(&addressBalance).Error

	case "withdraw":
		// complete the withdraw request and remove the event amount from the locked balance
		lockedBalance := ConvertStringToBigFloat(addressBalance.LockedBalance)
		amount := ConvertStringToBigFloat(addressEvent.Amount)
		lockedBalance.Sub(lockedBalance, amount)

		addressBalance.LockedBalance = lockedBalance.String()
		return tx.Save(&addressBalance).Error
	}
	return nil
}

func ConvertStringToBigFloat(num string) *big.Float {
	value, valid := new(big.Float).SetString(num)
	if !valid {
		return new(big.Float)
	}
	return value
}