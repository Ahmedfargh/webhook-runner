package presenter

import (
	pb "webhookRunner/api/proto/v1"
	"webhookRunner/internal/models"
)

type WebhookPresenter interface {
	ToProto(call *models.WebhookCall) *pb.WebhookCall
	ToProtoList(calls []models.WebhookCall) []*pb.WebhookCall
}

type webhookPresenter struct{}

func NewWebhookPresenter() WebhookPresenter {
	return &webhookPresenter{}
}

func (p *webhookPresenter) ToProto(call *models.WebhookCall) *pb.WebhookCall {
	if call == nil {
		return nil
	}
	appName := ""
	if call.App.Name != "" {
		appName = call.App.Name
	}

	return &pb.WebhookCall{
		Id:                 call.ID.String(),
		AppId:              call.AppID.String(),
		AppName:            appName,
		EventName:          call.EventName,
		TargetUrl:          call.TargetURL,
		PayloadJson:        call.PayloadJSON,
		HeadersJson:        call.HeadersJSON,
		AttemptCount:       call.AttemptCount,
		Status:             string(call.Status),
		ResponseStatusCode: call.ResponseStatusCode,
		ResponseBody:       call.ResponseBody,
		ResponseLatencyMs:  call.ResponseLatencyMS,
		ErrorMessage:       call.ErrorMessage,
		CreatedAt:          call.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          call.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (p *webhookPresenter) ToProtoList(calls []models.WebhookCall) []*pb.WebhookCall {
	result := make([]*pb.WebhookCall, len(calls))
	for i := range calls {
		result[i] = p.ToProto(&calls[i])
	}
	return result
}
