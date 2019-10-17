package repository

import (
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/stretchr/testify/mock"
)

// MockRepository contains additional methods for inspection
type MockRepository struct {
	mock.Mock
}

// CreateSubscription create a subscription
func (m *MockRepository) CreateSubscription(subscriptionContext *models.SubscriptionContext) (*models.SubscriptionContext, error) {
	rets := m.Called(subscriptionContext.Subscription.CallbackURL)
	/* Since `rets.Get()` is a generic method, that returns whatever we pass to it,
	 * we need to typecast it to the type we expect, which in this case is []*subscription
	 */
	return subscriptionContext, rets.Error(1)
}

// ReadSubscription read a subscription
func (m *MockRepository) ReadSubscription(id string) (*models.SubscriptionContext, error) {
	rets := m.Called(id)
	/* Since `rets.Get()` is a generic method, that returns whatever we pass to it,
	 * we need to typecast it to the type we expect, which in this case is []*subscription
	 */
	return rets.Get(0).(*models.SubscriptionContext), rets.Error(1)
}

// UpdateSubscription update a subscription
func (m *MockRepository) UpdateSubscription(subscriptionContext *models.SubscriptionContext) (*models.SubscriptionContext, error) {
	rets := m.Called(subscriptionContext.Subscription.ID)
	/* Since `rets.Get()` is a generic method, that returns whatever we pass to it,
	 * we need to typecast it to the type we expect, which in this case is []*subscription
	 */
	return subscriptionContext, rets.Error(1)
}

// DeleteSubscription delete a subscription
func (m *MockRepository) DeleteSubscription(id string) error {
	/* When this method is called, `m.Called` records the call, and also returns the result that we pass to it
	 * (which you will see in the handler tests)
	 */
	// rets := m.Called(fmt.Sprintf("DeleteSubscription(%s)", id))
	rets := m.Called(id)
	return rets.Error(0)
}

// GetActiveSubscriptions get active subscriptions
func (m *MockRepository) GetActiveSubscriptions(eventType models.EventType) (models.SubscriptionContexts, error) {
	rets := m.Called(eventType)
	/* Since `rets.Get()` is a generic method, that returns whatever we pass to it,
	 * we need to typecast it to the type we expect, which in this case is []*subscription
	 */
	return rets.Get(0).(models.SubscriptionContexts), rets.Error(1)
}

// InitMockRepository initialize repository
func InitMockRepository() *MockRepository {
	/*
		Like the InitStore function we defined earlier, this function
		also initializes the store variable, but this time, it assigns
		a new MockRepository instance to it, instead of an actual store
	*/
	s := new(MockRepository)
	SubscriptionRepository = s
	return s
}
