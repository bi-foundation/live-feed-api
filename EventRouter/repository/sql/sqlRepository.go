package sql

import (
	"database/sql"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

const (
	createSubscriptionQuery = "INSERT INTO subscriptions (callback, callback_type) VALUES(?, ?);"
	readSubscriptionQuery   = "SELECT id, id, callback, callback_type FROM subscription WHERE id = ?"
	readSubscriptionsQuery  = "SELECT id, id, callback, callback_type FROM subscription"
	deleteSubscriptionQuery = "INSERT INTO subscriptions (callback, callback_type) VALUES(?, ?);"
	updateSubscriptionQuery = "INSERT INTO subscriptions (callback, callback_type) VALUES(?, ?);"
)

var connection *sql.DB

type sqlRepository struct {
	createStatement *sql.Stmt
	readStatement   *sql.Stmt
	updateStatement *sql.Stmt
	deleteStatement *sql.Stmt
}

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

		connection = db
		log.Info("sql repository connected to: %s", version)
	}

	return repository, repository.prepareStatements()
}

func (repository *sqlRepository) prepareStatements() error {
	createStatement, err := connection.Prepare("INSERT INTO subscriptions (callback, callback_type) VALUES(?, ?);")
	if err != nil {
		return fmt.Errorf("failed to create subscription statement: %v", err)
	}

	readStatement, err := connection.Prepare("SELECT id, callback, callback_type FROM subscriptions WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to create subscription statement: %v", err)
	}

	updateStatement, err := connection.Prepare("UPDATE subscriptions SET callback = ? , callback_type = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to update subscription statement: %v", err)
	}

	deleteStatement, err := connection.Prepare("DELETE FROM subscriptions WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to delete subscription statement: %v", err)
	}

	repository.createStatement = createStatement
	repository.readStatement = readStatement
	repository.updateStatement = updateStatement
	repository.deleteStatement = deleteStatement
	return nil
}

func (repository *sqlRepository) Close() error {
	log.Info("closing connection")
	return connection.Close()
}

func (repository *sqlRepository) CreateSubscription(subscription *models.Subscription) (*models.Subscription, error) {
	createSql := "INSERT INTO subscriptions (callback, callback_type) VALUES(?, ?);"
	filterSql := "INSERT INTO filters (subscription, event_type, filtering) VALUES(?, ?, ?);"
	tx, err := connection.Begin()
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
	subscriptionStmt, err := tx.Prepare(createSql)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription statement: %v", err)
	}

	filterStmt, err := tx.Prepare(filterSql)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription statement: %v", err)
	}

	// insert subscription
	result, err := subscriptionStmt.Exec(subscription.Callback, subscription.CallbackType)
	rows, err := result.RowsAffected()
	if rows != 1 || err != nil {
		return nil, fmt.Errorf("failed to create subscription: no subscription found")
	}

	// extract inserted subscription id
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %v", err)
	}
	subscription.Id = strconv.FormatInt(id, 10)

	// insert filters
	for eventType, filter := range subscription.Filters {
		if _, err = filterStmt.Exec(id, eventType, filter.Filtering); err != nil {
			return nil, fmt.Errorf("failed to create subscription filter: %v", err)
		}
	}

	log.Debug("stored subscription: %v", subscription)
	return subscription, nil
}

func (repository *sqlRepository) ReadSubscription(id string) (*models.Subscription, error) {
	query := "SELECT callback, callback_type, event_type, filtering FROM subscriptions LEFT JOIN filters ON filters.subscription = subscriptions.id WHERE subscriptions.id = ?"
	rows, err := connection.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscription: %v", err)
	}

	subscription := &models.Subscription{
		Id:      id,
		Filters: make(map[models.EventType]models.Filter),
	}

	found := false
	for rows.Next() {
		found = true

		var eventTypeValue sql.NullString
		var filteringValue sql.NullString

		err := rows.Scan(&subscription.Callback, &subscription.CallbackType, &eventTypeValue, &filteringValue)
		if err != nil {
			return nil, fmt.Errorf("failed to read subscription: %v", err)
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
		return nil, fmt.Errorf("failed to read subscription: no subscriptions found with if '%s'", id)
	}

	log.Debug("read subscription: %v", subscription)
	return subscription, nil
}

func (repository *sqlRepository) UpdateSubscription(subscription *models.Subscription) (*models.Subscription, error) {
	updateSubscriptionQuery := "UPDATE subscriptions SET callback = ?, callback_type = ? WHERE id = ?"
	updateFilterQuery := "UPDATE filters SET filtering = ? WHERE subscription = ? AND event_type = ?"
	insertFilterQuery := "INSERT INTO filters(subscription, event_type, filtering) VALUES (?, ?, ?)"
	deleteFilterQuery := "DELETE FROM filters WHERE subscription = ? AND event_type = ?"

	oldSubscription, err := repository.ReadSubscription(subscription.Id)
	if err != nil {
		return nil, err
	}

	tx, err := connection.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription transaction: %v", err)
	}

	// commit or rollback when there is an error
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	if subscription.Callback != oldSubscription.Callback || subscription.CallbackType != oldSubscription.CallbackType {
		_, err := tx.Query(updateSubscriptionQuery, subscription.Callback, subscription.CallbackType, subscription.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to update subscription: %v", err)
		}
	}

	oldFilters := oldSubscription.Filters
	for eventType, filter := range subscription.Filters {
		// update existing filter or insert new filter
		if oldFilter, ok := oldFilters[eventType]; ok {
			// change update filtering, otherwise nothing changed
			if oldFilter.Filtering != filter.Filtering {
				_, err := tx.Query(updateFilterQuery, filter.Filtering, subscription.Id, eventType)
				if err != nil {
					return nil, fmt.Errorf("failed to update subscription filter: %v", err)
				}
			}

			// keep track of filter such that removed filter can be deleted from the db
			delete(oldFilters, eventType)
		} else {
			_, err := tx.Query(insertFilterQuery, filter.Filtering, subscription.Id, eventType)
			if err != nil {
				return nil, fmt.Errorf("failed to update subscription new filter: %v", err)
			}
		}
	}

	for eventType, _ := range oldFilters {
		_, err := tx.Query(deleteFilterQuery, subscription.Id, eventType)
		if err != nil {
			return nil, fmt.Errorf("failed to update subscription removed filter: %v", err)
		}
	}

	log.Debug("update subscription: %v", subscription)
	return subscription, nil
}

func (repository *sqlRepository) DeleteSubscription(id string) error {
	deleteFiltersQuery := "DELETE FROM filters WHERE subscription = ?"
	deleteSubscriptionsQuery := "DELETE FROM subscriptions WHERE id = ?"

	tx, err := connection.Begin()
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %v", err)
	}

	// commit or rollback when there is an error
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.Query(deleteFiltersQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %v", err)
	}

	_, err = tx.Query(deleteSubscriptionsQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %v", err)
	}

	log.Debug("deleted subscription: %s", id)
	return nil
}

func (repository *sqlRepository) GetSubscriptions(eventType models.EventType) ([]*models.Subscription, error) {
	subs := make(map[string]*models.Subscription)

	query := "SELECT subscription, event_type, filtering, callback, callback_type FROM filters LEFT JOIN subscriptions ON subscription = subscriptions.id"
	rows, err := connection.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %v", err)
	}

	for rows.Next() {
		var eventType models.EventType
		subscription := &models.Subscription{
			Filters: make(map[models.EventType]models.Filter),
		}
		filter := models.Filter{}

		err := rows.Scan(&subscription.Id, &eventType, &filter.Filtering, &subscription.Callback, &subscription.CallbackType)
		if err != nil {
			return nil, fmt.Errorf("failed to get subscriptions: %v", err)
		}

		if _, ok := subs[subscription.Id]; !ok {
			subs[subscription.Id] = subscription
		}

		subs[subscription.Id].Filters[eventType] = filter
	}

	// preallocate memory for the slice
	subscriptions := make([]*models.Subscription, 0, len(subs))
	for _, p := range subs {
		subscriptions = append(subscriptions, p)
	}

	log.Debug("get subscriptions: %v", subscriptions)
	return subscriptions, nil
}

func (repository *sqlRepository) GetAllSubscriptions() ([]*models.Subscription, error) {
	readSubscriptionsQuery := "SELECT id, id, callback, callback_type FROM subscriptions"
	subs := make(map[int]*models.Subscription)
	rows, err := connection.Query(readSubscriptionsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions statement: %v", err)
	}

	for rows.Next() {
		var id int
		subscription := &models.Subscription{
			Filters: map[models.EventType]models.Filter{},
		}
		rows.Scan(&id, &subscription.Id, &subscription.Callback, &subscription.CallbackType)
		subs[id] = subscription
	}

	readFiltersQuery := "SELECT subscription, event_type, filtering  FROM filters"
	rows, err = connection.Query(readFiltersQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get filters statement: %v", err)
	}

	for rows.Next() {
		var subscription int
		var eventType models.EventType
		filter := &models.Filter{}
		rows.Scan(&subscription, &eventType, &filter.Filtering)

		subs[subscription].Filters[eventType] = *filter
	}

	// preallocate memory for the slice
	subscriptions := make([]*models.Subscription, 0, len(subs))
	for _, p := range subs {
		subscriptions = append(subscriptions, p)
	}

	return subscriptions, nil
}
