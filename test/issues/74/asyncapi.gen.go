// Package "issue74" provides primitives to interact with the AsyncAPI specification.
//
// Code generated by github.com/lerenn/asyncapi-codegen version (devel) DO NOT EDIT.
package issue74

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lerenn/asyncapi-codegen/pkg/extensions"
)

// AppSubscriber represents all handlers that are expecting messages for App
type AppSubscriber interface {
	// TestChannel subscribes to messages placed on the 'testChannel' channel
	TestChannel(ctx context.Context, msg TestMessage)
}

// AppController is the structure that provides publishing capabilities to the
// developer and and connect the broker with the App
type AppController struct {
	controller
}

// NewAppController links the App to the broker
func NewAppController(bc extensions.BrokerController, options ...ControllerOption) (*AppController, error) {
	// Check if broker controller has been provided
	if bc == nil {
		return nil, extensions.ErrNilBrokerController
	}

	// Create default controller
	controller := controller{
		broker:        bc,
		subscriptions: make(map[string]extensions.BrokerChannelSubscription),
		logger:        extensions.DummyLogger{},
		middlewares:   make([]extensions.Middleware, 0),
	}

	// Apply options
	for _, option := range options {
		option(&controller)
	}

	return &AppController{controller: controller}, nil
}

func (c AppController) wrapMiddlewares(middlewares []extensions.Middleware, last extensions.NextMiddleware) func(ctx context.Context) {
	var called bool

	// If there is no more middleware
	if len(middlewares) == 0 {
		return func(ctx context.Context) {
			if !called {
				called = true
				last(ctx)
			}
		}
	}

	// Wrap middleware into a check function that will call execute the middleware
	// and call the next wrapped middleware if the returned function has not been
	// called already
	next := c.wrapMiddlewares(middlewares[1:], last)
	return func(ctx context.Context) {
		// Call the middleware and the following if it has not been done already
		if !called {
			called = true
			ctx = middlewares[0](ctx, next)

			// If next has already been called in middleware, it should not be
			// executed again
			next(ctx)
		}
	}
}

func (c AppController) executeMiddlewares(ctx context.Context, callback func(ctx context.Context)) {
	// Wrap middleware to have 'next' function when calling them
	wrapped := c.wrapMiddlewares(c.middlewares, callback)

	// Execute wrapped middlewares
	wrapped(ctx)
}

func addAppContextValues(ctx context.Context, path string) context.Context {
	ctx = context.WithValue(ctx, extensions.ContextKeyIsVersion, "1.0.0")
	ctx = context.WithValue(ctx, extensions.ContextKeyIsProvider, "app")
	return context.WithValue(ctx, extensions.ContextKeyIsChannel, path)
}

// Close will clean up any existing resources on the controller
func (c *AppController) Close(ctx context.Context) {
	// Unsubscribing remaining channels
	c.UnsubscribeAll(ctx)

	c.logger.Info(ctx, "Closed app controller")
}

// SubscribeAll will subscribe to channels without parameters on which the app is expecting messages.
// For channels with parameters, they should be subscribed independently.
func (c *AppController) SubscribeAll(ctx context.Context, as AppSubscriber) error {
	if as == nil {
		return extensions.ErrNilAppSubscriber
	}

	if err := c.SubscribeTestChannel(ctx, as.TestChannel); err != nil {
		return err
	}

	return nil
}

// UnsubscribeAll will unsubscribe all remaining subscribed channels
func (c *AppController) UnsubscribeAll(ctx context.Context) {
	c.UnsubscribeTestChannel(ctx)
}

// SubscribeTestChannel will subscribe to new messages from 'testChannel' channel.
//
// Callback function 'fn' will be called each time a new message is received.
func (c *AppController) SubscribeTestChannel(ctx context.Context, fn func(ctx context.Context, msg TestMessage)) error {
	// Get channel path
	path := "testChannel"

	// Set context
	ctx = addAppContextValues(ctx, path)
	ctx = context.WithValue(ctx, extensions.ContextKeyIsMessageDirection, "reception")

	// Check if there is already a subscription
	_, exists := c.subscriptions[path]
	if exists {
		err := fmt.Errorf("%w: %q channel is already subscribed", extensions.ErrAlreadySubscribedChannel, path)
		c.logger.Error(ctx, err.Error())
		return err
	}

	// Subscribe to broker channel
	sub, err := c.broker.Subscribe(ctx, path)
	if err != nil {
		c.logger.Error(ctx, err.Error())
		return err
	}
	c.logger.Info(ctx, "Subscribed to channel")

	// Asynchronously listen to new messages and pass them to app subscriber
	go func() {
		for {
			// Wait for next message
			bMsg, open := <-sub.MessagesChannel()

			// If subscription is closed and there is no more message
			// (i.e. uninitialized message), then exit the function
			if !open && bMsg.IsUninitialized() {
				return
			}

			// Set broker message to context
			ctx = context.WithValue(ctx, extensions.ContextKeyIsBrokerMessage, bMsg)

			// Process message
			msg, err := newTestMessageFromBrokerMessage(bMsg)
			if err != nil {
				c.logger.Error(ctx, err.Error())
			}
			msgCtx := context.WithValue(ctx, extensions.ContextKeyIsMessage, msg)

			// Execute middlewares with the callback
			c.executeMiddlewares(msgCtx, func(ctx context.Context) {
				fn(ctx, msg)
			})
		}
	}()

	// Add the cancel channel to the inside map
	c.subscriptions[path] = sub

	return nil
}

// UnsubscribeTestChannel will unsubscribe messages from 'testChannel' channel.
// A timeout can be set in context to avoid blocking operation, if needed.
func (c *AppController) UnsubscribeTestChannel(ctx context.Context) {
	// Get channel path
	path := "testChannel"

	// Check if there subscribers for this channel
	sub, exists := c.subscriptions[path]
	if !exists {
		return
	}

	// Set context
	ctx = addAppContextValues(ctx, path)

	// Stop the subscription
	sub.Cancel(ctx)

	// Remove if from the subscribers
	delete(c.subscriptions, path)

	c.logger.Info(ctx, "Unsubscribed from channel")
}

// UserController is the structure that provides publishing capabilities to the
// developer and and connect the broker with the User
type UserController struct {
	controller
}

// NewUserController links the User to the broker
func NewUserController(bc extensions.BrokerController, options ...ControllerOption) (*UserController, error) {
	// Check if broker controller has been provided
	if bc == nil {
		return nil, extensions.ErrNilBrokerController
	}

	// Create default controller
	controller := controller{
		broker:        bc,
		subscriptions: make(map[string]extensions.BrokerChannelSubscription),
		logger:        extensions.DummyLogger{},
		middlewares:   make([]extensions.Middleware, 0),
	}

	// Apply options
	for _, option := range options {
		option(&controller)
	}

	return &UserController{controller: controller}, nil
}

func (c UserController) wrapMiddlewares(middlewares []extensions.Middleware, last extensions.NextMiddleware) func(ctx context.Context) {
	var called bool

	// If there is no more middleware
	if len(middlewares) == 0 {
		return func(ctx context.Context) {
			if !called {
				called = true
				last(ctx)
			}
		}
	}

	// Wrap middleware into a check function that will call execute the middleware
	// and call the next wrapped middleware if the returned function has not been
	// called already
	next := c.wrapMiddlewares(middlewares[1:], last)
	return func(ctx context.Context) {
		// Call the middleware and the following if it has not been done already
		if !called {
			called = true
			ctx = middlewares[0](ctx, next)

			// If next has already been called in middleware, it should not be
			// executed again
			next(ctx)
		}
	}
}

func (c UserController) executeMiddlewares(ctx context.Context, callback func(ctx context.Context)) {
	// Wrap middleware to have 'next' function when calling them
	wrapped := c.wrapMiddlewares(c.middlewares, callback)

	// Execute wrapped middlewares
	wrapped(ctx)
}

func addUserContextValues(ctx context.Context, path string) context.Context {
	ctx = context.WithValue(ctx, extensions.ContextKeyIsVersion, "1.0.0")
	ctx = context.WithValue(ctx, extensions.ContextKeyIsProvider, "user")
	return context.WithValue(ctx, extensions.ContextKeyIsChannel, path)
}

// Close will clean up any existing resources on the controller
func (c *UserController) Close(ctx context.Context) {
	// Unsubscribing remaining channels
}

// PublishTestChannel will publish messages to 'testChannel' channel
func (c *UserController) PublishTestChannel(ctx context.Context, msg TestMessage) error {
	// Get channel path
	path := "testChannel"

	// Set context
	ctx = addUserContextValues(ctx, path)
	ctx = context.WithValue(ctx, extensions.ContextKeyIsMessage, msg)
	ctx = context.WithValue(ctx, extensions.ContextKeyIsMessageDirection, "publication")

	// Convert to BrokerMessage
	bMsg, err := msg.toBrokerMessage()
	if err != nil {
		return err
	}

	// Set broker message to context
	ctx = context.WithValue(ctx, extensions.ContextKeyIsBrokerMessage, bMsg)

	// Publish the message on event-broker through middlewares
	c.executeMiddlewares(ctx, func(ctx context.Context) {
		err = c.broker.Publish(ctx, path, bMsg)
	})

	// Return error from publication on broker
	return err
}

// controller is the controller that will be used to communicate with the broker
// It will be used internally by AppController and UserController
type controller struct {
	// broker is the broker controller that will be used to communicate
	broker extensions.BrokerController
	// subscriptions is a map of all subscriptions
	subscriptions map[string]extensions.BrokerChannelSubscription
	// logger is the logger that will be used² to log operations on controller
	logger extensions.Logger
	// middlewares are the middlewares that will be executed when sending or
	// receiving messages
	middlewares []extensions.Middleware
}

// ControllerOption is the type of the options that can be passed
// when creating a new Controller
type ControllerOption func(controller *controller)

// WithLogger attaches a logger to the controller
func WithLogger(logger extensions.Logger) ControllerOption {
	return func(controller *controller) {
		controller.logger = logger
	}
}

// WithMiddlewares attaches middlewares that will be executed when sending or receiving messages
func WithMiddlewares(middlewares ...extensions.Middleware) ControllerOption {
	return func(controller *controller) {
		controller.middlewares = middlewares
	}
}

type MessageWithCorrelationID interface {
	CorrelationID() string
	SetCorrelationID(id string)
}

type Error struct {
	Channel string
	Err     error
}

func (e *Error) Error() string {
	return fmt.Sprintf("channel %q: err %v", e.Channel, e.Err)
}

// TestMessage is the message expected for 'Test' channel
// test message
type TestMessage struct {
	// Headers will be used to fill the message headers
	Headers HeaderSchema

	// Payload will be inserted in the message payload
	Payload struct {
		Obj1 struct {
			// Description: reference ID.
			ReferenceID string `json:"reference_id"`
		} `json:"obj1"`
	}
}

func NewTestMessage() TestMessage {
	var msg TestMessage

	return msg
}

// newTestMessageFromBrokerMessage will fill a new TestMessage with data from generic broker message
func newTestMessageFromBrokerMessage(bMsg extensions.BrokerMessage) (TestMessage, error) {
	var msg TestMessage

	// Unmarshal payload to expected message payload format
	err := json.Unmarshal(bMsg.Payload, &msg.Payload)
	if err != nil {
		return msg, err
	}

	// Get each headers from broker message
	for k, v := range bMsg.Headers {
		switch {
		case k == "dateTime": // Retrieving DateTime header
			t, err := time.Parse(time.RFC3339, string(v))
			if err != nil {
				return msg, err
			}
			msg.Headers.DateTime = t
		case k == "version": // Retrieving Version header
			msg.Headers.Version = string(v)
		default:
			// TODO: log unknown error
		}
	}

	// TODO: run checks on msg type

	return msg, nil
}

// toBrokerMessage will generate a generic broker message from TestMessage data
func (msg TestMessage) toBrokerMessage() (extensions.BrokerMessage, error) {
	// TODO: implement checks on message

	// Marshal payload to JSON
	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return extensions.BrokerMessage{}, err
	}

	// Add each headers to broker message
	headers := make(map[string][]byte, 2)

	// Adding DateTime header
	headers["dateTime"] = []byte(msg.Headers.DateTime.Format(time.RFC3339)) // Adding Version header
	headers["version"] = []byte(msg.Headers.Version)

	return extensions.BrokerMessage{
		Headers: headers,
		Payload: payload,
	}, nil
}

// HeaderSchema is a schema from the AsyncAPI specification required in messages
// Description: header
type HeaderSchema struct {
	// Description: Date in UTC format "YYYY-MM-DDThh:mm:ss.sZ".
	DateTime time.Time `json:"date_time"`

	// Description: Schema version
	Version string `json:"version"`
}

// TestSchemaSchema is a schema from the AsyncAPI specification required in messages
type TestSchemaSchema struct {
	Obj1 struct {
		// Description: reference ID.
		ReferenceID string `json:"reference_id"`
	} `json:"obj1"`
}
