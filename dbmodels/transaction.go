package dbmodels

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	CashTransactionType = 0
	CardTransactionType = 1
)

type Transaction struct {
	Id bson.ObjectId `bson:"_id,omitempty" json:"id"`

	PayerId    bson.ObjectId `bson:"payerId,omitempty" json:"payerId"`
	ReceiverId bson.ObjectId `bson:"receiverId,omitempty" json:"receiverId"`

	Type     int       `bson:"type,omitempty" json:"type"`
	Ammount  float32   `bson:"ammount,omitempty" json:"ammount"`
	Currency string    `bson:"currency,omitempty" json:"currency"`
	Date     time.Time `bson:"date,omitempty" json:"date"`
}

func (transaction *Transaction) Equal(otherTransaction Transaction) bool {
	switch {
	case transaction.Id != otherTransaction.Id:
		return false
	case transaction.PayerId != otherTransaction.PayerId:
		return false
	case transaction.ReceiverId != otherTransaction.ReceiverId:
		return false
	case transaction.Type != otherTransaction.Type:
		return false
	case transaction.Ammount != otherTransaction.Ammount:
		return false
	case transaction.Currency != otherTransaction.Currency:
		return false
	case !transaction.Date.Truncate(time.Millisecond).Equal(otherTransaction.Date.Truncate(time.Millisecond)):
		return false
	}

	return true
}
