package postgres

import (
	"database/sql"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	_ "github.com/lib/pq"
	"strconv"
)

const (
	selectSubscriptionSql   = `SELECT callback, callback_type, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = $1;`
	selectSubscriptionsSql  = `SELECT subscription, callback, callback_type, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE event_type = $1;`
	insertSubscriptionSql   = `INSERT INTO subscriptions (callback, callback_type) VALUES($1, $2);`
	insertFilterSql         = `INSERT INTO filters (subscription, event_type, filtering) VALUES($1, $2, $3);`
	updateSubscriptionQuery = `UPDATE subscriptions SET callback = $1, callback_type = $2 WHERE id = $3;`
	updateFilterQuery       = `UPDATE filters SET filtering = $1 WHERE subscription = $2 AND event_type = $3;`
	deleteFilterSql         = `DELETE FROM filters WHERE subscription = $1 AND event_type = $2;`
	deleteFiltersSql        = `DELETE FROM filters WHERE subscription = $1;`
	deleteSubscriptionsSql  = `DELETE FROM subscriptions WHERE id = $1;`
)

var postgresConnection *sql.DB

type postgresRepository struct{}

func NewPostgres() (*postgresRepository, error) {
	repository := &postgresRepository{}
	return repository.connect()
}

func (repository *postgresRepository) connect() (*postgresRepository, error) {
	// open new connection if connection is nil or not open (if there is such a state)
	// you can also check "once.Do" if that suits your needs better
	if postgresConnection == nil {
		// TODO make configurable: driverName, user, password, url

		user := "postgres"
		password := "dbPassword"
		name := "live_api"
		connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, name)

		db, err := sql.Open("postgres", connStr)
		//db, err := sql.Open("mysql", "live-api:dbPassword@tcp(:3306)/test")
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

		postgresConnection = db
		log.Info("sql repository connected to: %s", version)
	}

	return repository, nil
}

func (repository *postgresRepository) Close() error {
	log.Info("closing connection")
	return postgresConnection.Close()
}

func (repository *postgresRepository) CreateSubscription(insertSubscription *models.Subscription) (subscription *models.Subscription, err error) {
	tx, err := postgresConnection.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription transaction: %v", err)
	}
	// commit or rollback when there is an error
	defer func() {
		if err != nil {
			tx.Rollback()
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
	result, err := subscriptionStmt.Exec(insertSubscription.Callback, insertSubscription.CallbackType)
	rows, err := result.RowsAffected()
	if rows != 1 || err != nil {
		err = fmt.Errorf("failed to create subscription: no subscription found")
		return nil, err
	}

	// extract inserted subscription id
	id, err := result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("failed to create subscription: %v", err)
		return nil, err
	}

	subscription = insertSubscription
	subscription.Id = strconv.FormatInt(id, 10)

	if len(insertSubscription.Filters) > 0 {
		filterStmt, err := tx.Prepare(insertFilterSql)
		if err != nil {
			err = fmt.Errorf("failed to create subscription statement: %v", err)
			return nil, err
		}

		// insert filters
		for eventType, filter := range insertSubscription.Filters {
			if _, err = filterStmt.Exec(id, eventType, filter.Filtering); err != nil {
				err = fmt.Errorf("failed to create subscription filter: %v", err)
				return nil, err
			}
		}
	}
	log.Debug("stored subscription: %v", subscription)
	return subscription, err
}

func (repository *postgresRepository) ReadSubscription(id string) (subscription *models.Subscription, err error) {
	rows, err := postgresConnection.Query(selectSubscriptionSql, id)
	if err != nil {
		err = fmt.Errorf("failed to read subscription: %v", err)
		return nil, err
	}

	subscription = &models.Subscription{
		Id:      id,
		Filters: make(map[models.EventType]models.Filter),
	}

	found := false
	for rows.Next() {
		found = true

		var eventTypeValue sql.NullString
		var filteringValue sql.NullString

		err = rows.Scan(&subscription.Callback, &subscription.CallbackType, &eventTypeValue, &filteringValue)
		if err != nil {
			err = fmt.Errorf("failed to read subscription: %v", err)
			return nil, err
		}

		if eventTypeValue.Valid {
			filter := models.Filter{}
			if filteringValue.Valid {
				filter.Filtering = models.GraphQL(filteringValue.String)
			}
			eventType := models.EventType(eventTypeValue.String)
			subscription.Filters[eventType] = filter
		}
	}

	if !found {
		err = fmt.Errorf("failed to read subscription: no subscriptions found with if '%s'", id)
		return nil, err
	}

	log.Debug("read subscription: %v", subscription)
	return subscription, err
}

func (repository *postgresRepository) UpdateSubscription(updateSubscription *models.Subscription) (subscription *models.Subscription, err error) {
	oldSubscription, err := repository.ReadSubscription(updateSubscription.Id)
	if err != nil {
		return nil, err
	}

	tx, err := postgresConnection.Begin()
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

	if updateSubscription.Callback != oldSubscription.Callback || updateSubscription.CallbackType != oldSubscription.CallbackType {
		_, err = tx.Exec(updateSubscriptionQuery, updateSubscription.Callback, updateSubscription.CallbackType, updateSubscription.Id)
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

	subscription = updateSubscription
	log.Debug("update subscription: %v", subscription)
	return subscription, err
}

func (repository *postgresRepository) DeleteSubscription(id string) (err error) {
	tx, err := postgresConnection.Begin()
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

func (repository *postgresRepository) GetSubscriptions(eventType models.EventType) (subscriptions []*models.Subscription, err error) {
	subs := make(map[string]*models.Subscription)

	rows, err := postgresConnection.Query(selectSubscriptionsSql, eventType)
	if err != nil {
		err = fmt.Errorf("failed to get subscriptions: %v", err)
		return nil, err
	}

	for rows.Next() {
		var eventType models.EventType
		subscription := &models.Subscription{
			Filters: make(map[models.EventType]models.Filter),
		}
		filter := models.Filter{}

		var eventTypeValue sql.NullString
		var filteringValue sql.NullString

		err = rows.Scan(&subscription.Id, &subscription.Callback, &subscription.CallbackType, &eventTypeValue, &filteringValue)
		if err != nil {
			err = fmt.Errorf("failed to get subscriptions: %v", err)
			return nil, err
		}

		if eventTypeValue.Valid {
			filter := models.Filter{}
			if filteringValue.Valid {
				filter.Filtering = models.GraphQL(filteringValue.String)
			}
			eventType := models.EventType(eventTypeValue.String)
			subscription.Filters[eventType] = filter
		}

		if _, ok := subs[subscription.Id]; !ok {
			subs[subscription.Id] = subscription
		}

		subs[subscription.Id].Filters[eventType] = filter
	}

	// preallocate memory for the slice
	subscriptions = make([]*models.Subscription, 0, len(subs))
	for _, p := range subs {
		subscriptions = append(subscriptions, p)
	}

	log.Debug("get subscriptions: %v", subscriptions)
	return subscriptions, err
}
