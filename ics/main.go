package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-playground/validator/v10"

	c "github.com/quesurifn/ics-calendar-tidbyt-lambda/ics/calendar"
	t "github.com/quesurifn/ics-calendar-tidbyt-lambda/ics/types"
	"go.uber.org/zap"
)

type Response events.APIGatewayProxyResponse

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return GetErrorResponseType(err)
	}

	event := &t.IcsRequest{}
	err = json.Unmarshal([]byte(request.Body), event)
	if err != nil {
		logger.Error("Error", zap.Any("err", err))
		return GetErrorResponseType(err)
	}

	val := validator.New()
	err = val.Struct(event)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		logger.Error("Error", zap.Any("err", validationErrors))
		return GetErrorResponseType(validationErrors)
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
		return GetErrorResponseType(err)
	}

	events, err := cal.ParseCalendar(data, event.TZ)
	if err != nil {
		logger.Error("Error", zap.Any("err", err))
		return GetErrorResponseType(err)
	}

	if len(events) == 0 {
		base := t.BaseResponse{
			Data: nil,
		}
		respBytes, err := json.Marshal(base)
		if err != nil {
			logger.Error("Error", zap.Any("err", err))
			return GetErrorResponseType(err)
		}

		return Response{
			StatusCode:      200,
			IsBase64Encoded: false,
			Body:            string(respBytes),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	nextEvent, err := cal.NextEvent(events, event.TZ)
	if err != nil {
		logger.Error("Error", zap.Any("err", err))
		return GetErrorResponseType(err)
	}

	response := t.BaseResponse{
		Data: &t.IcsResponse{
			Name:      nextEvent.Name,
			StartTime: nextEvent.StartTime,
			EndTime:   nextEvent.EndTime,
			Location:  nextEvent.Location,
			Detail: &t.EventDetail{
				IsToday:           nextEvent.Detail.IsToday,
				IsTomorrow:        nextEvent.Detail.IsTomorrow,
				IsThisWeek:        nextEvent.Detail.IsThisWeek,
				MinutesUntilStart: nextEvent.Detail.MinutesUntilStart,
				MinutesUntilEnd:   nextEvent.Detail.MinutesUntilEnd,
				HoursToEnd:        nextEvent.Detail.HoursToEnd,
				InProgress:        nextEvent.Detail.InProgress,
				IsAllDay:		   nextEvent.Detail.IsAllDay,
			},
		},
	}

	respBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("Error", zap.Any("err", err))
		return GetErrorResponseType(err)
	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(respBytes),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func GetErrorResponseType(err error) (Response, error) {
	resp := &t.BaseResponse{
		Error: &t.ErrorResponse{
			Error:   true,
			Message: err.Error(),
		},
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return Response{
			StatusCode: 500,
		}, fmt.Errorf("failed to marshal response: %w", err)
	}

	formattedResp := Response{
		StatusCode:      400,
		IsBase64Encoded: false,
		Body:            string(respBytes),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return formattedResp, nil
}
