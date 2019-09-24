package repository

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/FactomProject/live-feed-api/EventRouter/config"
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

	dbConfig := &config.DatabaseConfig{
		Database:         "",
		ConnectionString: "",
	}
	repository, _ := NewSQLRepository(dbConfig) // Ignore connection errors when using mocked connection
	repo, _ := repository.(*sqlRepository)
	connection = db
	return repo, mock
}

// test read subscription
func TestReadSubscription(t *testing.T) {
	repository, mock := initTest(t)

	// subscription to create
	subscription := models.Subscription{
		ID:           "1",
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BlockCommit:       {Filtering: fmt.Sprintf("filtering 1")},
			models.EntryRegistration: {Filtering: fmt.Sprintf("filtering 2")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.ID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.BlockCommit, subscription.Filters[models.BlockCommit].Filtering).
			AddRow(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.EntryRegistration, subscription.Filters[models.EntryRegistration].Filtering))

	// now we execute our methods
	readSubscriptionContext, err := repository.ReadSubscription(subscription.ID)
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
		ID:           "1",
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     1,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.ID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, nil, nil))

	// now we execute our methods
	readSubscriptionContext, err := repository.ReadSubscription(subscription.ID)
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
		CallbackURL:  "url",
		CallbackType: models.HTTP,
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO subscriptions \(failures, callback, callback_type, status, info, access_token, username, password\) VALUES\(\?, \?, \?, \?, \?, \?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword).WillReturnResult(sqlmock.NewResult(1, 1))
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
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BlockCommit: {Filtering: fmt.Sprintf("filtering 1")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}
	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO subscriptions \(failures, callback, callback_type, status, info, access_token, username, password\) VALUES\(\?, \?, \?, \?, \?, \?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(`INSERT INTO filters \(subscription, event_type, filtering\) VALUES\(\?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO filters`).WithArgs(1, models.BlockCommit, subscription.Filters[models.BlockCommit].Filtering).WillReturnResult(sqlmock.NewResult(1, 1))
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
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BlockCommit: {Filtering: fmt.Sprintf("filtering 1")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO subscriptions \(failures, callback, callback_type, status, info, access_token, username, password\) VALUES\(\?, \?, \?, \?, \?, \?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(`INSERT INTO filters \(subscription, event_type, filtering\) VALUES\(\?, \?, \?\);`)
	mock.ExpectExec(`INSERT INTO filters`).WithArgs(1, models.BlockCommit, subscription.Filters[models.BlockCommit].Filtering).
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
		ID:           "42",
		CallbackURL:  "url",
		CallbackType: models.HTTP,
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.ID).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, nil, nil))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.ID).WillReturnResult(sqlmock.NewResult(42, 1))
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
		ID:           "42",
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BlockCommit:              {Filtering: fmt.Sprintf("no change filtering")},
			models.EntryContentRegistration: {Filtering: fmt.Sprintf("insert new filtering")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.ID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.BlockCommit, "no change filtering"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.ID).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`INSERT INTO filters`).WithArgs("42", models.EntryContentRegistration, subscription.Filters[models.EntryContentRegistration].Filtering).WillReturnResult(sqlmock.NewResult(1, 1)).WillReturnResult(sqlmock.NewResult(42, 1))
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
		ID:           "42",
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BlockCommit:       {Filtering: fmt.Sprintf("no change filtering")},
			models.EntryRegistration: {Filtering: fmt.Sprintf("update this filtering")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.ID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.BlockCommit, "no change filtering").
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.EntryRegistration, "this will be changed"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.ID).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`UPDATE filters`).WithArgs(subscription.Filters[models.EntryRegistration].Filtering, "42", models.EntryRegistration).WillReturnResult(sqlmock.NewResult(1, 1)).WillReturnResult(sqlmock.NewResult(42, 1))
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
		ID:           "42",
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters: map[models.EventType]models.Filter{
			models.BlockCommit: {Filtering: fmt.Sprintf("no change filtering")},
		},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.ID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.BlockCommit, "no change filtering").
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.ChainRegistration, "this will be deleted"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.ID).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`DELETE FROM filters`).WithArgs(subscription.ID, models.ChainRegistration).WillReturnResult(sqlmock.NewResult(42, 1))
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
		ID:           "42",
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.ID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.EntryRegistration, "filtering"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).
		WithArgs(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.ID).
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
		ID:           "42",
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.ID).
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
		ID:           "42",
		CallbackURL:  "url",
		CallbackType: models.HTTP,
		Filters:      map[models.EventType]models.Filter{},
	}
	subscriptionContext := &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	columns := []string{"failures", "callback", "callback_type", "status", "info", "access_token", "username", "password", "event_type", "filtering"}
	mock.ExpectQuery(`SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = \?`).
		WithArgs(subscription.ID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(subscriptionContext.Failures, "url-change", subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, models.EntryRegistration, "this will be deleted"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE subscriptions`).WithArgs(subscriptionContext.Failures, subscription.CallbackURL, subscription.CallbackType, subscription.SubscriptionStatus, subscription.SubscriptionInfo, subscription.Credentials.AccessToken, subscription.Credentials.BasicAuthUsername, subscription.Credentials.BasicAuthPassword, subscription.ID).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(`DELETE FROM filters`).WithArgs(subscription.ID, models.EntryRegistration).WillReturnError(fmt.Errorf("some error"))
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
		WithArgs(models.BlockCommit).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, 0, "url", models.HTTP, models.Active, "", "", "", "", models.BlockCommit, "should be returned").
			AddRow(1, 0, "url", models.HTTP, models.Active, "", "", "", "", models.EntryRegistration, "should be returned").
			AddRow(2, 1, "url", models.HTTP, models.Active, "", "", "", "", nil, nil).
			AddRow(3, 2, "url", models.HTTP, models.Active, "", "", "", "", models.BlockCommit, "return"))

	// now we execute our methods
	subscriptionContexts, err := repository.GetActiveSubscriptions(models.BlockCommit)
	if err != nil {
		t.Errorf("error was not expected creating subscription: %s", err)
	}

	assert.Equal(t, 3, len(subscriptionContexts))

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
