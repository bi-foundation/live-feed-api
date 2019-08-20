package sql

import (
	"database/sql"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/models/errors"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

const (
	selectSubscriptionSql   = `SELECT failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = ?;`
	selectSubscriptionsSql  = `SELECT subscription, failures, callback, callback_type, status, info, access_token, username, password, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE event_type = ?;`
	insertSubscriptionSql   = `INSERT INTO subscriptions (failures, callback, callback_type, status, info, access_token, username, password) VALUES(?, ?, ?, ?, ?, ?, ?, ?);`
	insertFilterSql         = `INSERT INTO filters (subscription, event_type, filtering) VALUES(?, ?, ?);`
	updateSubscriptionQuery = `UPDATE subscriptions SET failures = ?, callback = ?, callback_type = ?, status = ?, info = ?, access_token = ?, username = ?, password = ? WHERE id = ?`
	updateFilterQuery       = `UPDATE filters SET filtering = ? WHERE subscription = ? AND event_type = ?`
	deleteFilterSql         = `DELETE FROM filters WHERE subscription = ? AND event_type = ?`
	deleteFiltersSql        = `DELETE FROM filters WHERE subscription = ?`
	deleteSubscriptionsSql  = `DELETE FROM subscriptions WHERE id = ?`
)

var connection *sql.DB

type sqlRepository struct{}

func New() (*sqlRepository, error) {
	repository := &sqlRepository{}
	return repository.connect()
}

func (repository *sqlRepository) connect() (*sqlRepository, error) {
	// open new connection if connection is nil or not open (if there is such a state)
	// you can also check "once.Do" if that suits your needs better
	if connection == nil {
		// TODO make configurable: driverName, user, password, url
		db, err := sql.Open("mysql", "live-api:jJBAGyB5MBhshzcC@tcp(127.0.0.1:3306)/live_api")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to sql database: %v", err)
		}

		err = db.Ping()
		if err != nil {
			return nil, fmt.Errorf("ping failed: %v", err)
		}

		// Connect and check the server version
		var version string
		err = db.QueryRow("SELECT VERSION()").Scan(&version)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to server for version: %v", err)
		}

		connection = db
		log.Info("sql repository connected to: %s", version)
	}

	return repository, nil
}

func (repository *sqlRepository) Close() error {
	log.Info("closing connection")
	return connection.Close()
}

func (repository *sqlRepository) CreateSubscription(createSubscriptionContext *models.SubscriptionContext) (subscriptionContext *models.SubscriptionContext, err error) {
	tx, err := connection.Begin()
	if err != nil {
		err = fmt.Errorf("failed to create subscription transaction: %v", err)
		return nil, err
	}
	// commit or rollback when there is an error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// prepare statements
	subscriptionStmt, err := tx.Prepare(insertSubscriptionSql)
	if err != nil {
		err = fmt.Errorf("failed to create subscription statement: %v", err)
		return nil, err
	}

	// insert subscription
	createSubscription := &createSubscriptionContext.Subscription
	result, err := subscriptionStmt.Exec(createSubscriptionContext.Failures, createSubscription.CallbackUrl, createSubscription.CallbackType, createSubscription.SubscriptionStatus, createSubscription.SubscriptionInfo, createSubscription.Credentials.AccessToken, createSubscription.Credentials.BasicAuthUsername, createSubscription.Credentials.BasicAuthPassword)
	if err != nil {
		err = fmt.Errorf("failed to create subscription: %v", err)
		return nil, err
	}
	rows, err := result.RowsAffected()
	if rows != 1 || err != nil {
		err = fmt.Errorf("failed to create subscription: %v", err)
		return nil, err
	}

	// extract inserted subscription id
	id, err := result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("failed to create subscription: %v", err)
		return nil, err
	}

	subscription := *createSubscription
	subscription.Id = strconv.FormatInt(id, 10)
	subscriptionContext = &models.SubscriptionContext{
		Subscription: subscription,
		Failures:     0,
	}

	if len(createSubscription.Filters) > 0 {
		filterStmt, err := tx.Prepare(insertFilterSql)
		if err != nil {
			err = fmt.Errorf("failed to create subscription statement: %v", err)
			return nil, err
		}

		// insert filters
		for eventType, filter := range createSubscription.Filters {
			if _, err = filterStmt.Exec(id, eventType, filter.Filtering); err != nil {
				err = fmt.Errorf("failed to create subscription filter: %v", err)
				return nil, err
			}
		}
	}
	log.Debug("stored subscription: %v", subscriptionContext)
	return subscriptionContext, err
}

func (repository *sqlRepository) ReadSubscription(id string) (subscriptionContext *models.SubscriptionContext, err error) {
	rows, err := connection.Query(selectSubscriptionSql, id)
	if err != nil {
		err = fmt.Errorf("failed to read subscription: %v", err)
		return nil, err
	}

	subscriptionContext = &models.SubscriptionContext{
		Subscription: models.Subscription{
			Id:      id,
			Filters: make(map[models.EventType]models.Filter),
		},
		Failures: 0,
	}
	subscription := &subscriptionContext.Subscription

	found := false
	for rows.Next() {
		found = true

		var eventTypeValue sql.NullString
		var filteringValue sql.NullString

		err = rows.Scan(&subscriptionContext.Failures, &subscription.CallbackUrl, &subscription.CallbackType, &subscription.SubscriptionStatus, &subscription.SubscriptionInfo, &subscription.Credentials.AccessToken, &subscription.Credentials.BasicAuthUsername, &subscription.Credentials.BasicAuthPassword, &eventTypeValue, &filteringValue)
		if err != nil {
			err = fmt.Errorf("failed to read subscription: %v", err)
			return nil, err
		}

		if eventTypeValue.Valid {
			filter := models.Filter{}
			if filteringValue.Valid {
				filter.Filtering = filteringValue.String
			}
			eventType := models.EventType(eventTypeValue.String)
			subscription.Filters[eventType] = filter
		}
	}

	if !found {
		return nil, errors.NewSubscriptionNotFound(id)
	}

	log.Debug("read subscription: %v", subscriptionContext)
	return subscriptionContext, err
}

func (repository *sqlRepository) UpdateSubscription(updateSubscriptionContext *models.SubscriptionContext) (subscriptionContext *models.SubscriptionContext, err error) {
	updateSubscription := &updateSubscriptionContext.Subscription
	oldSubscriptionContext, err := repository.ReadSubscription(updateSubscription.Id)
	if err != nil {
		return nil, err
	}

	tx, err := connection.Begin()
	if err != nil {
		err = fmt.Errorf("failed to update subscription transaction: %v", err)
		return nil, err
	}

	// commit or rollback when there is an error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// check if the subscription needs to be updated
	oldSubscription := &oldSubscriptionContext.Subscription
	if updateSubscriptionContext.Failures != oldSubscriptionContext.Failures ||
		updateSubscription.CallbackUrl != oldSubscription.CallbackUrl ||
		updateSubscription.CallbackType != oldSubscription.CallbackType ||
		updateSubscription.SubscriptionStatus != oldSubscription.SubscriptionStatus ||
		updateSubscription.SubscriptionInfo != oldSubscription.SubscriptionInfo ||
		updateSubscription.Credentials.AccessToken != oldSubscription.Credentials.AccessToken ||
		updateSubscription.Credentials.BasicAuthUsername != oldSubscription.Credentials.BasicAuthUsername ||
		updateSubscription.Credentials.BasicAuthPassword != oldSubscription.Credentials.BasicAuthPassword {

		_, err = tx.Exec(updateSubscriptionQuery, updateSubscriptionContext.Failures, updateSubscription.CallbackUrl, updateSubscription.CallbackType, updateSubscription.SubscriptionStatus, updateSubscription.SubscriptionInfo, updateSubscription.Credentials.AccessToken, updateSubscription.Credentials.BasicAuthUsername, updateSubscription.Credentials.BasicAuthPassword, updateSubscription.Id)
		if err != nil {
			err = fmt.Errorf("failed to update subscription: %v", err)
			return nil, err
		}
	}

	oldFilters := oldSubscription.Filters
	for eventType, filter := range updateSubscription.Filters {
		// update existing filter or insert new filter
		if oldFilter, ok := oldFilters[eventType]; ok {
			// change update filtering, otherwise nothing changed
			if oldFilter.Filtering != filter.Filtering {
				_, err = tx.Exec(updateFilterQuery, filter.Filtering, updateSubscription.Id, eventType)
				if err != nil {
					err = fmt.Errorf("failed to update subscription filter: %v", err)
					return nil, err
				}
			}

			// keep track of filter such that removed filter can be deleted from the db
			delete(oldFilters, eventType)
		} else {
			_, err = tx.Exec(insertFilterSql, updateSubscription.Id, eventType, filter.Filtering)
			if err != nil {
				err = fmt.Errorf("failed to update subscription new filter: %v", err)
				return nil, err
			}
		}
	}

	for eventType := range oldFilters {
		_, err = tx.Exec(deleteFilterSql, updateSubscription.Id, eventType)
		if err != nil {
			err = fmt.Errorf("failed to update subscription removed filter: %v", err)
			return nil, err
		}
	}

	subscriptionContext = updateSubscriptionContext
	log.Debug("update subscription: %v", subscriptionContext)
	return subscriptionContext, err
}

func (repository *sqlRepository) DeleteSubscription(id string) (err error) {
	tx, err := connection.Begin()
	if err != nil {
		err = fmt.Errorf("failed to delete subscription: %v", err)
		return err
	}

	// commit or rollback when there is an error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.Exec(deleteFiltersSql, id)
	if err != nil {
		err = fmt.Errorf("failed to delete subscription: %v", err)
		return err
	}

	_, err = tx.Exec(deleteSubscriptionsSql, id)
	if err != nil {
		err = fmt.Errorf("failed to delete subscription: %v", err)
		return err
	}

	log.Debug("deleted subscription: %s", id)
	return err
}

func (repository *sqlRepository) GetSubscriptions(eventType models.EventType) (subscriptionContexts []*models.SubscriptionContext, err error) {
	subContexts := make(map[string]*models.SubscriptionContext)

	rows, err := connection.Query(selectSubscriptionsSql, eventType)
	if err != nil {
		err = fmt.Errorf("failed to get subscriptions: %v", err)
		return nil, err
	}

	for rows.Next() {
		var eventType models.EventType
		subscriptionContext := &models.SubscriptionContext{
			Subscription: models.Subscription{
				Filters: make(map[models.EventType]models.Filter),
			},
		}
		subscription := &subscriptionContext.Subscription
		filter := models.Filter{}

		var eventTypeValue sql.NullString
		var filteringValue sql.NullString

		err = rows.Scan(&subscription.Id, &subscriptionContext.Failures, &subscription.CallbackUrl, &subscription.CallbackType, &subscription.SubscriptionStatus, &subscription.SubscriptionInfo, &subscription.Credentials.AccessToken, &subscription.Credentials.BasicAuthUsername, &subscription.Credentials.BasicAuthPassword, &eventTypeValue, &filteringValue)
		if err != nil {
			err = fmt.Errorf("failed to get subscriptions: %v", err)
			return nil, err
		}

		if eventTypeValue.Valid {
			filter := models.Filter{}
			if filteringValue.Valid {
				filter.Filtering = filteringValue.String
			}
			eventType := models.EventType(eventTypeValue.String)
			subscription.Filters[eventType] = filter
		}

		if _, ok := subContexts[subscription.Id]; !ok {
			subContexts[subscription.Id] = subscriptionContext
		}

		subContexts[subscription.Id].Subscription.Filters[eventType] = filter
	}

	// preallocate memory for the slice
	subscriptionContexts = make([]*models.SubscriptionContext, 0, len(subContexts))
	for _, s := range subContexts {
		subscriptionContexts = append(subscriptionContexts, s)
	}

	log.Debug("get subscriptions: %v", subscriptionContexts)
	return subscriptionContexts, err
}
