package sql

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
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

	repository, err := New()
	if err != nil {
		t.Fatalf("failed to create repository: %v\n", err)
	}

	connection = db
	return repository, mock
}

// test read subscription
func TestReadSubscription(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := &models.Subscription{
		Id:           "1",
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("filtering 1"))},
			models.COMMIT_ENTRY: {Filtering: models.GraphQL(fmt.Sprintf("filtering 2"))},
		},
	}

	columns := []string{"callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.ANCHOR_EVENT, subscription.Filters[models.ANCHOR_EVENT].Filtering).
			AddRow(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.COMMIT_ENTRY, subscription.Filters[models.COMMIT_ENTRY].Filtering))

	// now we execute our methods
	readSubscription, err := repository.ReadSubscription(subscription.Id)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscription, readSubscription)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test read subscription
func TestReadSubscriptionWithFilterNil(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := &models.Subscription{
		Id:           "1",
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}

	columns := []string{"callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, nil, nil))

	// now we execute our methods
	readSubscription, err := repository.ReadSubscription(subscription.Id)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscription, readSubscription)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test insert subscription
func TestCreateSubscription(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := &models.Subscription{
		Callback:     "url",
		CallbackType: models.HTTP,
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO subscriptions \(callback, callback_type, status, info, access_token, username, password\) VALUES\(\?, \?, \?, \?, \?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO subscriptions`).WithArgs(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// now we execute our method
	createdSubscription, err := repository.CreateSubscription(subscription)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscription, createdSubscription)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test insert subscription with filters
func TestCreateSubscriptionAddFilter(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := &models.Subscription{
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("filtering 1"))},
		},
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO subscriptions \(callback, callback_type, status, info, access_token, username, password\) VALUES\(\?, \?, \?, \?, \?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO subscriptions`).WithArgs(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(`INSERT INTO filters \(subscription, event_type, filtering\) VALUES\(\?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO filters`).WithArgs(1, models.ANCHOR_EVENT, subscription.Filters[models.ANCHOR_EVENT].Filtering).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// now we execute our method
	if _, err := repository.CreateSubscription(subscription); err != nil {
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
	subscription := &models.Subscription{
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("filtering 1"))},
		},
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO subscriptions \(callback, callback_type, status, info, access_token, username, password\) VALUES\(\?, \?, \?, \?, \?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO subscriptions`).WithArgs(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(`INSERT INTO filters \(subscription, event_type, filtering\) VALUES\(\?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO filters`).WithArgs(1, models.ANCHOR_EVENT, subscription.Filters[models.ANCHOR_EVENT].Filtering).
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	// now we execute our method
	if _, err := repository.CreateSubscription(subscription); err == nil {
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
	subscription := &models.Subscription{
		Id:           "42",
		Callback:     "url",
		CallbackType: models.HTTP,
	}

	columns := []string{"callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).AddRow("url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, nil, nil))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()

	// now we execute our method
	updatedSubscription, err := repository.UpdateSubscription(subscription)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscription, updatedSubscription)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test update subscription add one filter to the existing filters
func TestUpdateSubscriptionAddFilter(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := &models.Subscription{
		Id:           "42",
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("no change filtering"))},
			models.REVEAL_ENTRY: {Filtering: models.GraphQL(fmt.Sprintf("insert new filtering"))},
		},
	}

	columns := []string{"callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.ANCHOR_EVENT, "no change filtering"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`INSERT INTO filters`).WithArgs("42", models.REVEAL_ENTRY, subscription.Filters[models.REVEAL_ENTRY].Filtering).WillReturnResult(sqlmock.NewResult(1, 1)).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()

	// now we execute our method
	updatedSubscription, err := repository.UpdateSubscription(subscription)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscription, updatedSubscription)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test update subscription with updating a filter
func TestUpdateSubscriptionUpdateFilter(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := &models.Subscription{
		Id:           "42",
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("no change filtering"))},
			models.COMMIT_ENTRY: {Filtering: models.GraphQL(fmt.Sprintf("update this filtering"))},
		},
	}

	columns := []string{"callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.ANCHOR_EVENT, "no change filtering").
			AddRow("url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.COMMIT_ENTRY, "this will be changed"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`UPDATE filters`).WithArgs(subscription.Filters[models.COMMIT_ENTRY].Filtering, "42", models.COMMIT_ENTRY).WillReturnResult(sqlmock.NewResult(1, 1)).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()

	// now we execute our method
	updatedSubscription, err := repository.UpdateSubscription(subscription)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscription, updatedSubscription)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test update subscription and delete one filter
func TestUpdateSubscriptionDeleteFilter(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := &models.Subscription{
		Id:           "42",
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.ANCHOR_EVENT: {Filtering: models.GraphQL(fmt.Sprintf("no change filtering"))},
		},
	}

	columns := []string{"callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.ANCHOR_EVENT, "no change filtering").
			AddRow("url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.COMMIT_CHAIN, "this will be deleted"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`DELETE FROM filters`).WithArgs(subscription.Id, models.COMMIT_CHAIN).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()

	// now we execute our method
	updatedSubscription, err := repository.UpdateSubscription(subscription)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assertSubscription(t, subscription, updatedSubscription)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test a rollback on update subscription when updating the subscription
func TestUpdateSubscriptionRollbackOnUpdateFailure(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := &models.Subscription{
		Id:           "42",
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}

	columns := []string{"callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.COMMIT_ENTRY, "filtering"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).
		WithArgs(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	// now we execute our method
	if _, err := repository.UpdateSubscription(subscription); err == nil {
		t.Errorf("was expecting an error, but there was none")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// test a rollback on update subscription when delete a removed subscription
func TestUpdateSubscriptionRollbackOnDeleteFailure(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to update
	subscription := &models.Subscription{
		Id:           "42",
		Callback:     "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}

	columns := []string{"callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.Id).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.COMMIT_ENTRY, "this will be deleted"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscription.Callback, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.Id).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`DELETE FROM filters`).WithArgs(subscription.Id, models.COMMIT_ENTRY).WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	// now we execute our method
	if _, err := repository.UpdateSubscription(subscription); err == nil {
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

func TestGetSubscriptions(t *testing.T) {
	repository, mock := initTest(t)

	columns := []string{"subscription", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT subscription, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE event_type = \?`).
		WithArgs(models.ANCHOR_EVENT).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "url", models.HTTP, models.ACTIVE, "", "", "", "", models.ANCHOR_EVENT, "should be returned").
			AddRow(1, "url", models.HTTP, models.ACTIVE, "", "", "", "", models.COMMIT_ENTRY, "should be returned").
			AddRow(2, "url", models.HTTP, models.ACTIVE, "", "", "", "", nil, nil).
			AddRow(3, "url", models.HTTP, models.ACTIVE, "", "", "", "", models.ANCHOR_EVENT, "return"))

	// now we execute our methods
	subscriptions, err := repository.GetSubscriptions(models.ANCHOR_EVENT)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assert.Equal(t, 3, len(subscriptions))

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func assertSubscription(t *testing.T, expected *models.Subscription, actual *models.Subscription) {
	if actual == nil {
		assert.Fail(t, "subscription is nil")
		return
	}
	assert.NotNil(t, actual.Id)
	assert.Equal(t, expected.Callback, actual.Callback)
	assert.Equal(t, expected.CallbackType, actual.CallbackType)
	assert.Equal(t, expected.SubscriptionStatus, actual.SubscriptionStatus)
	assert.Equal(t, expected.SubscriptionInfo, actual.SubscriptionInfo)
	assert.Equal(t, expected.Credentials.AccessToken, actual.Credentials.AccessToken)
	assert.Equal(t, expected.Credentials.BasicAuthUsername, actual.Credentials.BasicAuthUsername)
	assert.Equal(t, expected.Credentials.BasicAuthPassword, actual.Credentials.BasicAuthPassword)
	assert.Equal(t, len(expected.Filters), len(actual.Filters))

	for eventType, filter := range expected.Filters {
		assert.NotNil(t, actual.Filters[eventType])
		assert.Equal(t, filter.Filtering, actual.Filters[eventType].Filtering)
	}
}
