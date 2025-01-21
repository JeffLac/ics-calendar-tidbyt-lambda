package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-playground/validator/v10"
	c "github.com/quesurifn/ics-calendar-tidbyt-lambda/calendar"
	t "github.com/quesurifn/ics-calendar-tidbyt-lambda/types"
	"go.uber.org/zap"
)

func HandleRequest(ctx context.Context, event *t.IcsRequest) (*t.BaseResponse, error) {
	logger, err := zap.NewProduction()
	if err != nil {

		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	if event == nil {
		logger.Error("event is nil")
		resp := GetErrorResponseType(errors.New("event is nil"))
		return resp, fmt.Errorf("event is nil")
	}

	val := validator.New()
	err = val.Struct(event)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		logger.Error("Error", zap.Any("err", validationErrors))
		resp := GetErrorResponseType(validationErrors)
		return resp, err
	}

	cal := c.Calendar{
		Logger: logger,
		TZMap: map[string]string{
			"Hawaii Standard Time":     "Pacific/Honolulu",
			"Alaskan Standard Time":    "America/Anchorage",
			"Alaskan Daylight Time":    "America/Anchorage",
			"SA Pacific Standard Time": "America/Bogota",
			"Pacific Standard Time":    "America/Los_Angeles",
			"Pacific Daylight Time":    "America/Los_Angeles",
			"Central Standard Time":    "America/Chicago",
			"Central Daylight Time":    "America/Chicago",
			"Mountain Standard Time":   "America/Denver",
			"Mountain Daylight Time":   "America/Denver",
			"Eastern Standard Time":    "America/New_York",
			"Eastern Daylight Time":    "America/New_York",
		},
	}

	data, err := cal.DownloadCalendar(event.ICSUrl)
	if err != nil {
		logger.Error("Error", zap.Any("err", err))
		resp := GetErrorResponseType(err)
		return resp, err
	}

	events, err := cal.ParseCalendar(data, event.TZ)
	if err != nil {
		logger.Error("Error", zap.Any("err", err))
		resp := GetErrorResponseType(err)
		return resp, err
	}

	//need to pass events and the TZ because there is an adjustment in NextEvent
	//for all day events
	nextEvent := cal.NextEvent(events, event.TZ)
	if err != nil {
		logger.Error("Error", zap.Any("err", err))
		resp := GetErrorResponseType(err)
		return resp, err
	}

	response := t.BaseResponse{
		Data: &t.IcsResponse{
			EventName:      nextEvent.Name,
			EventStartTime: nextEvent.StartTime,
			EventEndTime:   nextEvent.EndTime,
			EventLocation:  nextEvent.Location,
			InProgress:     nextEvent.InProgress,
		},
	}

	return &response, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func GetErrorResponseType(err error) *t.BaseResponse {
	return &t.BaseResponse{
		Error: &t.ErrorResponse{
			Error:   true,
			Message: err.Error(),
		},
	}
}