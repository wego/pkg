package errors

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/wego/pkg/common"
)

// CaptureError captures an error to sentry & set level as error
func CaptureError(ctx context.Context, err error) {
	hub := getHub(ctx)
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		enrichScope(ctx, scope, err)
		hub.CaptureException(err)
	})
}

// CaptureWarning captures an error to sentry & set level as warning
func CaptureWarning(ctx context.Context, err error) {
	hub := getHub(ctx)
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelWarning)
		enrichScope(ctx, scope, err)
		hub.CaptureException(err)
	})
}

func enrichScope(ctx context.Context, scope *sentry.Scope, err error) {
	errorCode := fmt.Sprint(Code(err))
	scope.SetTag("error_code", errorCode)
	fingerprint := []string{errorCode, err.Error()}

	e, ok := err.(*Error)
	if ok {
		clientCode := common.GetString(ctx, common.CtxClientCode)
		if len(clientCode) > 0 {
			scope.SetTag("client_code", clientCode)
			fingerprint = append(fingerprint, clientCode)
		}
		transRef := common.GetString(ctx, common.CtxTransactionRef)
		if len(transRef) > 0 {
			scope.SetTag("transaction_ref", transRef)
			fingerprint = append(fingerprint, transRef)
		}
		paymentRef := common.GetString(ctx, common.CtxPaymentRef)
		if len(paymentRef) > 0 {
			scope.SetTag("payment_ref", paymentRef)
			fingerprint = append(fingerprint, paymentRef)
		}
		deploymentRef := common.GetString(ctx, common.CtxDeploymentRef)
		if len(deploymentRef) > 0 {
			scope.SetTag("deployment_ref", deploymentRef)
			fingerprint = append(fingerprint, deploymentRef)
		}

		// extra is not searchable in sentry
		ops := ops(e)
		scope.SetExtra("operations", ops)
		for _, o := range ops {
			fingerprint = append(fingerprint, string(o))
		}

		extras := common.GetExtras(ctx)
		for k, v := range extras {
			scope.SetExtra(k, v)
		}
	}
	scope.SetFingerprint(fingerprint)

}

func getHub(ctx context.Context) (hub *sentry.Hub) {
	hub = sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	return
}
