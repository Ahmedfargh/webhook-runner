package presenter

import (
	pb "webhookRunner/api/proto/v1"
	"webhookRunner/internal/models"
)

type AppPresenter interface {
	ToProto(app *models.App) *pb.App
	ToProtoList(apps []models.App) []*pb.App
}

type appPresenter struct{}

func NewAppPresenter() AppPresenter {
	return &appPresenter{}
}

func (p *appPresenter) ToProto(app *models.App) *pb.App {
	if app == nil {
		return nil
	}
	return &pb.App{
		Id:            app.ID.String(),
		UserId:        app.UserID.String(),
		Name:          app.Name,
		AppId:         app.AppID,
		AppSecret:     app.AppSecret,
		WebhookUrl:    app.WebhookURL,
		WebhookSecret: app.WebhookSecret,
		IsActive:      app.IsActive,
		CreatedAt:     app.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     app.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (p *appPresenter) ToProtoList(apps []models.App) []*pb.App {
	result := make([]*pb.App, len(apps))
	for i := range apps {
		result[i] = p.ToProto(&apps[i])
	}
	return result
}
