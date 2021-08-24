package pubsubmessaging

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/common"
	"gitlab.slade360emr.com/go/commontools/crm/pkg/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ReceivePubSubPushMessages receives and processes a Pub/Sub push message.
func (ps ServicePubSubMessaging) ReceivePubSubPushMessages(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	message, err := ps.baseExt.VerifyPubSubJWTAndDecodePayload(w, r)
	if err != nil {
		ps.baseExt.WriteJSONResponse(
			w,
			ps.baseExt.ErrorMap(err),
			http.StatusBadRequest,
		)
		return
	}

	span.AddEvent("published message", trace.WithAttributes(
		attribute.Any("message", message),
	))

	topicID, err := ps.baseExt.GetPubSubTopic(message)
	if err != nil {
		ps.baseExt.WriteJSONResponse(
			w,
			ps.baseExt.ErrorMap(err),
			http.StatusBadRequest,
		)
		return
	}

	span.AddEvent("published message topic", trace.WithAttributes(
		attribute.String("topic", topicID),
	))

	switch topicID {
	case ps.AddPubSubNamespace(common.CreateCRMContact):
		var CRMContact domain.CRMContact
		err := json.Unmarshal(message.Message.Data, &CRMContact)
		if err != nil {
			ps.baseExt.WriteJSONResponse(
				w,
				ps.baseExt.ErrorMap(err),
				http.StatusBadRequest,
			)
			return
		}
		if _, err = ps.crm.CreateHubSpotContact(ctx, &CRMContact); err != nil {
			ps.baseExt.WriteJSONResponse(
				w,
				ps.baseExt.ErrorMap(err),
				http.StatusBadRequest,
			)
			return
		}

	case ps.AddPubSubNamespace(common.UpdateCRMContact):
		var CRMContact domain.CRMContact
		err := json.Unmarshal(message.Message.Data, &CRMContact)
		if err != nil {
			ps.baseExt.WriteJSONResponse(
				w,
				ps.baseExt.ErrorMap(err),
				http.StatusBadRequest,
			)
			return
		}
		if _, err = ps.crm.UpdateHubSpotContact(ctx, &CRMContact); err != nil {
			ps.baseExt.WriteJSONResponse(
				w,
				ps.baseExt.ErrorMap(err),
				http.StatusBadRequest,
			)
			return
		}

	default:
		errMsg := fmt.Sprintf(
			"pub sub handler error: unknown topic `%s`",
			topicID,
		)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	resp := map[string]string{"status": "success"}
	marshalledSuccessMsg, err := json.Marshal(resp)
	if err != nil {
		ps.baseExt.WriteJSONResponse(
			w,
			ps.baseExt.ErrorMap(err),
			http.StatusInternalServerError,
		)
		return
	}
	_, _ = w.Write(marshalledSuccessMsg)
}
