package transactionservice

import (
	"gopkg.in/mgo.v2/bson"
	"gost/config"
	"gost/dbmodels"
	"gost/service"
	"testing"
	"time"
)

func TestTransactionCRUD(t *testing.T) {
	transaction := &dbmodels.Transaction{}

	setUpTransactionsTest(t)
	defer tearDownTransactionsTest(t, transaction)

	createTransaction(t, transaction)
	verifyTransactionCorresponds(t, transaction)

	if !t.Failed() {
		changeAndUpdateTransaction(t, transaction)
		verifyTransactionCorresponds(t, transaction)
	}
}

func setUpTransactionsTest(t *testing.T) {
	config.InitTestsDatabase()
	service.InitDbService()

	if recover() != nil {
		t.Fatal("Test setup failed!")
	}
}

func tearDownTransactionsTest(t *testing.T, transaction *dbmodels.Transaction) {
	err := DeleteTransaction(transaction.Id)

	if err != nil {
		t.Fatal("The transaction document could not be deleted!")
	}
}

func createTransaction(t *testing.T, transaction *dbmodels.Transaction) {
	*transaction = dbmodels.Transaction{
		Id:         bson.NewObjectId(),
		PayerId:    bson.NewObjectId(),
		ReceiverId: bson.NewObjectId(),
		Type:       dbmodels.CASH_TRANSACTION_TYPE,
		Ammount:    6469.1264,
		Currency:   "RON",
		Date:       time.Now().Local(),
	}

	err := CreateTransaction(transaction)

	if err != nil {
		t.Fatal("The transaction document could not be created!")
	}
}

func changeAndUpdateTransaction(t *testing.T, transaction *dbmodels.Transaction) {
	transaction.PayerId = bson.NewObjectId()
	transaction.ReceiverId = bson.NewObjectId()
	transaction.Type = dbmodels.CARD_TRANSACTION_TYPE
	transaction.Currency = "USD"

	err := UpdateTransaction(transaction)

	if err != nil {
		t.Fatal("The transaction document could not be updated!")
	}
}

func verifyTransactionCorresponds(t *testing.T, transaction *dbmodels.Transaction) {
	dbtransaction, err := GetTransaction(transaction.Id)

	if err != nil || dbtransaction == nil {
		t.Error("Could not fetch the transaction document from the database!")
	}

	if !dbtransaction.Equal(transaction) {
		t.Error("The transaction document doesn't correspond with the document extracted from the database!")
	}
}
