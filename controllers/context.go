package controllers

import "context"

// UserMessages is used to pass UI messages between Handlers
// used only with context
// for UI messages within the same Handler, user ad hoc map[string]interface{} called d for data
type UserMessages map[string]interface{}
type key int

const userMessagesKey key = 99

var (
	ctx    context.Context
	cancel context.CancelFunc
)

// newContextUserM returns a new Context carrying userMessages
func newContextUserM(ctx context.Context, userM UserMessages) context.Context {
	return context.WithValue(ctx, userMessagesKey, userM)
}

// fromContextUserM retrieves userMessages from a http Request Context
func fromContextUserM(ctx context.Context) (UserMessages, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the userMessages type assertion returns ok=false for nil.
	userM, ok := ctx.Value(userMessagesKey).(UserMessages)
	return userM, ok
}
