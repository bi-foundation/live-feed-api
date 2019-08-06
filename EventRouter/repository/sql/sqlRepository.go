package sql

import (
	"database/sql"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

type Repository struct {
	db              *sql.DB
	createStatement *sql.Stmt
	readStatement   *sql.Stmt
	updateStatement *sql.Stmt
	deleteStatement *sql.Stmt
}

func NewSQLRepository() (repository.Repository, error) {
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

	log.Info("sql repository connected to: %s", version)

	createStatement, err := db.Prepare("INSERT INTO subscriptions(callback) VALUES (?)")
	if err != nil {
		return nil, fmt.Errorf("failed to create create statement: %v", err)
	}

	readStatement, err := db.Prepare("SELECT id, callback FROM subscriptions WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to create read statement: %v", err)
	}

	updateStatement, err := db.Prepare("UPDATE subscriptions SET callback = ? WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to update read statement: %v", err)
	}

	deleteStatement, err := db.Prepare("DELETE FROM subscriptions WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to delete read statement: %v", err)
	}

	repository := &Repository{
		db:              db,
		createStatement: createStatement,
		readStatement:   readStatement,
		updateStatement: updateStatement,
		deleteStatement: deleteStatement,
	}

	return repository, nil
}

func (repository *Repository) Close() error {
	defer repository.createStatement.Close()
	defer repository.readStatement.Close()
	return repository.db.Close()
}

func (repository *Repository) CreateSubscription(subscription *models.Subscription) (*models.Subscription, error) {
	result, err := repository.createStatement.Exec(subscription.Callback)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %v", err)
	}

	rows, err := result.RowsAffected()
	if rows != 1 || err != nil {
		return nil, fmt.Errorf("failed to create subscription: no subscription found")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %v", err)
	}
	subscription.Id = strconv.FormatInt(id, 10)

	log.Debug("stored subscription: %v", subscription)
	return subscription, nil
}

func (repository *Repository) ReadSubscription(id string) (*models.Subscription, error) {
	subscription := &models.Subscription{}
	err := repository.readStatement.QueryRow(id).Scan(&subscription.Id, &subscription.Callback)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscription: %v", err)
	}

	log.Debug("read subscription: %v", subscription)
	return subscription, nil
}

func (repository *Repository) UpdateSubscription(id string, subscription *models.Subscription) (*models.Subscription, error) {
	result, err := repository.updateStatement.Exec(subscription.Callback, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %v", err)
	}

	rows, err := result.RowsAffected()
	if rows != 1 || err != nil {
		return nil, fmt.Errorf("failed to create subscription: no subscription found")
	}

	log.Debug("update subscription: %v", subscription)
	return subscription, nil
}

func (repository *Repository) DeleteSubscription(id string) (*models.Subscription, error) {
	subscription, err := repository.ReadSubscription(id)
	if err != nil {
		return nil, err
	}

	result, err := repository.deleteStatement.Exec(id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete subscription: %v", err)
	}

	rows, err := result.RowsAffected()
	if rows != 1 || err != nil {
		return nil, fmt.Errorf("failed to delete subscription: no row has been updated")
	}

	log.Debug("deleted subscription: %v", id)
	return subscription, nil
}
