package controllers

import (
	"context"
	"errors"

	"github.com/nicomo/abacaxi/logger"
)

type key int

const (
	pageKey      key = 1
	tsnameKey    key = 2
	cFeedbackKey key = 3
)

var (
	ctx    context.Context
	cancel context.CancelFunc
)

// newContextPage returns a new Context with the page to skip to.
// Used when we want to display long lists of records
func newContextPage(ctx context.Context, p int) context.Context {
	return context.WithValue(ctx, pageKey, p)
}

// fromContextPage retrieves the page to skip to when paginating, from a http Request Context
func fromContextPage(ctx context.Context) (string, int, bool) {
	page, ok1 := ctx.Value(pageKey).(int)
	tsname, ok2 := ctx.Value(tsnameKey).(string)
	if !ok1 || !ok2 {
		return tsname, page, false
	}

	return tsname, page, true
}

// newContextTSName
func newContextTSName(ctx context.Context, tsname string) context.Context {
	return context.WithValue(ctx, tsnameKey, tsname)
}

// newContextCountFeedback
func newContextCountFeedback(ctx context.Context, results chan CountFeedback) context.Context {
	logger.Debug.Println("in newContextCountFeedback")
	return context.WithValue(ctx, cFeedbackKey, results)
}

// fromContextCountFeedback
func fromContextCountFeedback(ctx context.Context) (chan CountFeedback, error) {
	logger.Debug.Println("in fromContextCountFeedback")
	counter, ok := ctx.Value(cFeedbackKey).(chan CountFeedback)
	if !ok {
		return counter, errors.New("could not retrieve context")
	}
	return counter, nil
}
