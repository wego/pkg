package logger

import (
	"go.uber.org/zap"
)

// RequestType ...
type RequestType string
type logType string
type contextKey string

// PartnerRequestType list
const (
	RequestTypeValidateOrder         RequestType = "ValidateOrder"
	RequestTypeUpdatePostAuthResult  RequestType = "UpdatePostAuthResult"
	RequestTypeCompleteOrder         RequestType = "CompleteOrder"
	RequestTypeCancelOrderByCustomer RequestType = "CancelOrderByCustomer"
	RequestTypeCancelOrderByMerchant RequestType = "CancelOrderByMerchant"
	RequestTypeClaimChargeback       RequestType = "ClaimChargeback"

	RequestTypeFetchAccessToken RequestType = "FetchAccessToken"

	RequestTypeGetPaymentOptions      RequestType = "GetPaymentOptions"
	RequestTypeAuthorizePayment       RequestType = "AuthorizePayment"
	RequestTypeVoidPayment            RequestType = "VoidPayment"
	RequestTypeCapturePayment         RequestType = "CapturePayment"
	RequestTypeRefundPayment          RequestType = "RefundPayment"
	RequestTypeGetPaymentDetails      RequestType = "GetPaymentDetails"
	RequestTypeGetPaymentActions      RequestType = "GetPaymentActions"
	RequestTypeGetDisputes            RequestType = "GetDisputes"
	RequestTypeProcessDisputes        RequestType = "ProcessDisputes"
	RequestTypeGetDisputeDetails      RequestType = "GetDisputeDetails"
	RequestTypeQueryBIN               RequestType = "QueryBIN"
	RequestTypeRequestToken           RequestType = "RequestToken"
	RequestTypeRedirectPost3DS        RequestType = "RedirectPost3DS"
	RequestTypeProcessPendingPayments RequestType = "ProcessPendingPayments"

	RequestTypeDeployVirtualCard               RequestType = "DeployVirtualCard"
	RequestTypeUpdateVirtualCardDeployment     RequestType = "UpdateVirtualCardDeployment"
	RequestTypeGetVirtualCardDeploymentDetails RequestType = "GetVirtualCardDeploymentDetails"
	RequestTypeGetVirtualCardActivities        RequestType = "GetVirtualCardActivities"
	RequestTypeCancelVirtualCardDeployment     RequestType = "CancelVirtualCardDeployment"
	RequestTypeGetCardDetailsPublic            RequestType = "GetCardDetailsPublic"
	RequestTypeGetCardDetails                  RequestType = "GetCardDetails"

	RequestTypeSendDisputeEvent RequestType = "SendDisputeEvent"
	RequestTypeSendPaymentEvent RequestType = "SendPaymentEvent"

	RequestTypeUnknown RequestType = "Unknown"
)

const (
	logDir                  = "log"
	ultronExFileName        = "ultronex.{{env}}.log"
	partnerRequestsFileName = "partner_requests.{{env}}.log"
	requestsFileName        = "requests.{{env}}.log"
	defaultReplacement      = "[Filtered by Wego]"

	logTypeUltronex       logType = "ultronEx"
	logTypePartnerRequest logType = "partnerRequest"
	logTypeRequest        logType = "request"

	contextKeyRequest     contextKey = "request"
	contextKeyRequestType contextKey = "requestType"
)

var (
	loggers          map[logType]*zap.Logger
	sensitiveHeaders = map[string]bool{
		"authorization":   true,
		"x-forter-siteid": true,
	}
)
