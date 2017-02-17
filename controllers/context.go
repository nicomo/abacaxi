package controllers

import "context"

// UserMessages is used to pass UI messages between Handlers
// used only with context
// for UI messages within the same Handler, user ad hoc map[string]interface{} called d for data
type UserMessages map[string]interface{}
type key int

const (
	userMessagesKey key = 0
	pageKey         key = 1
	tsnameKey       key = 2
)

var (
	ctx    context.Context
	cancel context.CancelFunc
)

// newContextPage returns a new Context with the page to skip to (see long lists of records)
func newContextPage(ctx context.Context, p int) context.Context {
	return context.WithValue(ctx, pageKey, p)
}

// newContextTSName
func newContextTSName(ctx context.Context, tsname string) context.Context {
	return context.WithValue(ctx, tsnameKey, tsname)
}

// newContextUserM returns a new Context carrying userMessages
func newContextUserM(ctx context.Context, userM UserMessages) context.Context {
	return context.WithValue(ctx, userMessagesKey, userM)
}

// fromContextPage retrieves the page to skip to from a http Request Context
func fromContextPage(ctx context.Context) (string, int, bool) {
	page, ok1 := ctx.Value(pageKey).(int)
	tsname, ok2 := ctx.Value(tsnameKey).(string)
	if !ok1 || !ok2 {
		return tsname, page, false
	}

	return tsname, page, true
}

// fromContextUserM retrieves userMessages from a http Request Context
func fromContextUserM(ctx context.Context) (UserMessages, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the userMessages type assertion returns ok=false for nil.
	userM, ok := ctx.Value(userMessagesKey).(UserMessages)
	return userM, ok
}
