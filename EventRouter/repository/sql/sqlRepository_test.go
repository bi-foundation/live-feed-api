package sql

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/FactomProject/live-feed-api/EventRouter/models/errors"
	_ "github.com/proullon/ramsql/driver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func initTest(t *testing.T) (*sqlRepository, sqlmock.Sqlmock) {
	log.SetLevel(log.D)

	// init mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection\n", err)
	}

	repository, _ := New()

	connection = db
	return repository, mock
}

// test read subscription
func TestReadSubscription(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := models.Subscription{
		Id:           "1",
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BLOCK_COMMIT:       {Filtering: fmt.Sprintf("filtering 1")},
			models.ENTRY_REGISTRATION: {Filtering: fmt.Sprintf("filtering 2")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.BLOCK_COMMIT, subscription.Filters[models.BLOCK_COMMIT].Filtering).
			AddRow(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.ENTRY_REGISTRATION, subscription.Filters[models.ENTRY_REGISTRATION].Filtering))

	// now we execute our methods
	readSubscriptionContext, err := repository.ReadSubscription(subscription.Id)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscriptionContext, readSubscriptionContext)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestReadSubscriptionUnknownId(t *testing.T) {
	repository, mock := initTest(t)

	id := "1"
	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows(columns))

	// now we execute our methods
	readSubscriptionContext, err := repository.ReadSubscription(id)
	if err == nil {
		t.Errorf("was expecting an error, but there was none")
	}

	assert.IsType(t, errors.SubscriptionNotFound{}, err)
	assert.Nil(t, readSubscriptionContext)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test read subscription
func TestReadSubscriptionWithFilterNil(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := models.Subscription{
		Id:           "1",
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     1,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, nil, nil))

	// now we execute our methods
	readSubscriptionContext, err := repository.ReadSubscription(subscription.Id)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscriptionContext, readSubscriptionContext)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test insert subscription
func TestCreateSubscription(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := models.Subscription{
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO subscriptions \(failures, callback, callback_type, status, info, access_token, username, password\) VALUES\(\?, \?, \?, \?, \?, \?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// now we execute our method
	createdSubscriptionContext, err := repository.CreateSubscription(subscriptionContext)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscriptionContext, createdSubscriptionContext)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test insert subscription with filters
func TestCreateSubscriptionAddFilter(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := models.Subscription{
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BLOCK_COMMIT: {Filtering: fmt.Sprintf("filtering 1")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}
	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO subscriptions \(failures, callback, callback_type, status, info, access_token, username, password\) VALUES\(\?, \?, \?, \?, \?, \?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(`INSERT INTO filters \(subscription, event_type, filtering\) VALUES\(\?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO filters`).WithArgs(1, models.BLOCK_COMMIT, subscription.Filters[models.BLOCK_COMMIT].Filtering).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// now we execute our method
	if _, err := repository.CreateSubscription(subscriptionContext); err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test a rollback on insert subscription
func TestCreateSubscriptionRollbackOnFailure(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := models.Subscription{
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BLOCK_COMMIT: {Filtering: fmt.Sprintf("filtering 1")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO subscriptions \(failures, callback, callback_type, status, info, access_token, username, password\) VALUES\(\?, \?, \?, \?, \?, \?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(`INSERT INTO filters \(subscription, event_type, filtering\) VALUES\(\?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO filters`).WithArgs(1, models.BLOCK_COMMIT, subscription.Filters[models.BLOCK_COMMIT].Filtering).
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	// now we execute our method
	if _, err := repository.CreateSubscription(subscriptionContext); err == nil {
		t.Errorf("was expecting an error, but there was none")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test update a subscription without filters
func TestUpdateSubscription(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := models.Subscription{
		Id:           "42",
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, nil, nil))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()

	// now we execute our method
	updatedSubscriptionContext, err := repository.UpdateSubscription(subscriptionContext)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscriptionContext, updatedSubscriptionContext)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test update subscription add one filter to the existing filters
func TestUpdateSubscriptionAddFilter(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := models.Subscription{
		Id:           "42",
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BLOCK_COMMIT:               {Filtering: fmt.Sprintf("no change filtering")},
			models.ENTRY_CONTENT_REGISTRATION: {Filtering: fmt.Sprintf("insert new filtering")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.BLOCK_COMMIT, "no change filtering"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`INSERT INTO filters`).WithArgs("42", models.ENTRY_CONTENT_REGISTRATION, subscription.Filters[models.ENTRY_CONTENT_REGISTRATION].Filtering).WillReturnResult(sqlmock.NewResult(1, 1)).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()

	// now we execute our method
	updatedSubscriptionContext, err := repository.UpdateSubscription(subscriptionContext)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscriptionContext, updatedSubscriptionContext)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test update subscription with updating a filter
func TestUpdateSubscriptionUpdateFilter(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := models.Subscription{
		Id:           "42",
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BLOCK_COMMIT:       {Filtering: fmt.Sprintf("no change filtering")},
			models.ENTRY_REGISTRATION: {Filtering: fmt.Sprintf("update this filtering")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.BLOCK_COMMIT, "no change filtering").
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.ENTRY_REGISTRATION, "this will be changed"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`UPDATE filters`).WithArgs(subscription.Filters[models.ENTRY_REGISTRATION].Filtering, "42", models.ENTRY_REGISTRATION).WillReturnResult(sqlmock.NewResult(1, 1)).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()

	// now we execute our method
	updatedSubscriptionContext, err := repository.UpdateSubscription(subscriptionContext)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscriptionContext, updatedSubscriptionContext)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test update subscription and delete one filter
func TestUpdateSubscriptionDeleteFilter(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := models.Subscription{
		Id:           "42",
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BLOCK_COMMIT: {Filtering: fmt.Sprintf("no change filtering")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.BLOCK_COMMIT, "no change filtering").
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.CHAIN_REGISTRATION, "this will be deleted"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`DELETE FROM filters`).WithArgs(subscription.Id, models.CHAIN_REGISTRATION).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()

	// now we execute our method
	updatedSubscriptionContext, err := repository.UpdateSubscription(subscriptionContext)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscriptionContext, updatedSubscriptionContext)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test a rollback on update subscription when updating the subscription
func TestUpdateSubscriptionRollbackOnUpdateFailure(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := models.Subscription{
		Id:           "42",
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.ENTRY_REGISTRATION, "filtering"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).
		WithArgs(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	// now we execute our method
	if _, err := repository.UpdateSubscription(subscriptionContext); err == nil {
		t.Errorf("was expecting an error, but there was none")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test a rollback on update subscription when updating the subscription
func TestUpdateSubscriptionUnkownId(t *testing.T) {
	repository, mock := initTest(t)

	subscription := models.Subscription{
		Id:           "42",
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns))

	// now we execute our method
	updateSubscription, err := repository.UpdateSubscription(subscriptionContext)
	if err == nil {
		t.Errorf("was expecting an error, but there was none")
	}

	assert.IsType(t, errors.SubscriptionNotFound{}, err)
	assert.Nil(t, updateSubscription)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test a rollback on update subscription when delete a removed subscription
func TestUpdateSubscriptionRollbackOnDeleteFailure(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := models.Subscription{
		Id:           "42",
		CallbackUrl:  "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.ENTRY_REGISTRATION, "this will be deleted"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackUrl, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`DELETE FROM filters`).WithArgs(subscription.Id, models.ENTRY_REGISTRATION).WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	// now we execute our method
	if _, err := repository.UpdateSubscription(subscriptionContext); err == nil {
		t.Errorf("was expecting an error, but there was none")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteSubscription(t *testing.T) {
	repository, mock := initTest(t)

	id := "42"

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM filters`).WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM subscriptions`).WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// now we execute our method
	if err := repository.DeleteSubscription(id); err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteSubscriptionRollbackOnFailure(t *testing.T) {
	repository, mock := initTest(t)

	id := "42"

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM filters`).WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM subscriptions`).WithArgs(id).WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	// now we execute our method
	err := repository.DeleteSubscription(id)
	if err == nil {
		t.Errorf("was expecting an error, but there was none")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetActiveSubscriptions(t *testing.T) {
	repository, mock := initTest(t)

	columns := []string{"subscription", "failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT subscription, failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE event_type = \? AND status = 'ACTIVE'`).
		WithArgs(models.BLOCK_COMMIT).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, 0, "url", models.HTTP, models.ACTIVE, "", "", "", "", models.BLOCK_COMMIT, "should be returned").
			AddRow(1, 0, "url", models.HTTP, models.ACTIVE, "", "", "", "", models.ENTRY_REGISTRATION, "should be returned").
			AddRow(2, 1, "url", models.HTTP, models.ACTIVE, "", "", "", "", nil, nil).
			AddRow(3, 2, "url", models.HTTP, models.ACTIVE, "", "", "", "", models.BLOCK_COMMIT, "return"))

	// now we execute our methods
	subscriptionContexts, err := repository.GetActiveSubscriptions(models.BLOCK_COMMIT)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assert.Equal(t, 3, len(subscriptionContexts))

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func assertSubscription(t *testing.T, expected *models.SubscriptionContext, actual *models.SubscriptionContext) {
	if actual == nil {
		assert.Fail(t, "subscription is nil")
		return
	}
	assert.NotNil(t, actual.Subscription.Id)
	assert.Equal(t, expected.Failures, actual.Failures)
	assert.Equal(t, expected.Subscription.CallbackUrl, actual.Subscription.CallbackUrl)
	assert.Equal(t, expected.Subscription.CallbackType, actual.Subscription.CallbackType)
	assert.Equal(t, expected.Subscription.SubscriptionStatus, actual.Subscription.SubscriptionStatus)
	assert.Equal(t, expected.Subscription.SubscriptionInfo, actual.Subscription.SubscriptionInfo)
	assert.Equal(t, expected.Subscription.Credentials.AccessToken, actual.Subscription.Credentials.AccessToken)
	assert.Equal(t, expected.Subscription.Credentials.BasicAuthUsername, actual.Subscription.Credentials.BasicAuthUsername)
	assert.Equal(t, expected.Subscription.Credentials.BasicAuthPassword, actual.Subscription.Credentials.BasicAuthPassword)
	assert.Equal(t, len(expected.Subscription.Filters), len(actual.Subscription.Filters))

	for eventType, filter := range expected.Subscription.Filters {
		assert.NotNil(t, actual.Subscription.Filters[eventType])
		assert.Equal(t, filter.Filtering, actual.Subscription.Filters[eventType].Filtering)
	}
}
