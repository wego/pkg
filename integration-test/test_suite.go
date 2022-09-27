package integration_test

import "context"

// TestSuite interface for integration testing of external components such as database or webservice
type TestSuite interface {
	// StartUp - initializes the external component, returns the connection if database
	StartUp() interface{}
	// CleanUp - clear test data or shutdown the external component
	CleanUp()

	// Create a record in db
	Create(ctx context.Context, v interface{}) error
	// Find a record in db using primary key ID
	FindByID(ctx context.Context, entity interface{}, id uint) error
}
