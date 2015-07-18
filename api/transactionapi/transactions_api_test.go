package transactionapi

import (
	"go-server-template/api"
	"go-server-template/dbmodels"
	"go-server-template/models"
	"go-server-template/service/transactionservice"
	"go-server-template/tests"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"testing"
)

const transactionsRoute = "[{\"id\": \"TransactionsRoute\", \"pattern\": \"/transactions\", \"handlers\": {\"DELETE\": \"DeleteTransaction\", \"GET\": \"GetTransaction\", \"POST\": \"PostTransaction\"}}]"
const apiPath = "/transactions"

type dummyTransaction struct {
	BadField string
}

func (transaction *dummyTransaction) PopConstrains() {}

func TestTransactionsApi(t *testing.T) {
	tests.InitializeServerConfigurations(transactionsRoute, new(TransactionsApi))

	testPostTransactionInBadFormat(t)
	testPostTransactionNotIntegral(t)
	id := testPostTransactionInGoodFormat(t)
	testGetTransactionWithInexistentIdInDB(t)
	testGetTransactionWithBadIdParam(t)
	testGetTransactionWithGoodIdParam(t, id)

	// Delete the created transaction
	transactionservice.DeleteTransaction(id)
}

func testGetTransactionWithInexistentIdInDB(t *testing.T) {
	params := url.Values{}
	params.Add("id", bson.NewObjectId().Hex())

	tests.PerformApiTestCall(apiPath, api.GET, http.StatusNotFound, params, nil, t)
}

func testGetTransactionWithBadIdParam(t *testing.T) {
	params := url.Values{}
	params.Add("id", "2as456fas4")

	tests.PerformApiTestCall(apiPath, api.GET, http.StatusBadRequest, params, nil, t)
}

func testGetTransactionWithGoodIdParam(t *testing.T, id bson.ObjectId) {
	params := url.Values{}
	params.Add("id", id.Hex())

	rw := tests.PerformApiTestCall(apiPath, api.GET, http.StatusOK, params, nil, t)

	body := rw.Body.String()
	if len(body) == 0 {
		t.Error("Response body is empty or in deteriorated format:", body)
	}
}

func testPostTransactionInBadFormat(t *testing.T) {
	dTransaction := &dummyTransaction{
		BadField: "bad value",
	}

	tests.PerformApiTestCall(apiPath, api.POST, http.StatusBadRequest, nil, dTransaction, t)
}

func testPostTransactionNotIntegral(t *testing.T) {
	transaction := &models.Transaction{
		Id:       bson.NewObjectId(),
		Payer:    models.User{Id: bson.NewObjectId()},
		Currency: "USD",
	}

	tests.PerformApiTestCall(apiPath, api.POST, http.StatusBadRequest, nil, transaction, t)
}

func testPostTransactionInGoodFormat(t *testing.T) bson.ObjectId {
	transaction := &models.Transaction{
		Id:       bson.NewObjectId(),
		Payer:    models.User{Id: bson.NewObjectId()},
		Receiver: models.User{Id: bson.NewObjectId()},
		Type:     dbmodels.CashTransactionType,
		Ammount:  216.365,
		Currency: "USD",
	}

	rw := tests.PerformApiTestCall(apiPath, api.POST, http.StatusCreated, nil, transaction, t)

	body := rw.Body.String()
	if len(body) == 0 {
		t.Error("Response body is empty or in deteriorated format:", body)
	}

	return transaction.Id
}
