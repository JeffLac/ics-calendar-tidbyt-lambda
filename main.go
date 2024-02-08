package main

import (
	"context"
	"errors"
	"fmt"
	"syscall"

	"github.com/aws/aws-lambda-go/lambda"
	c "github.com/quesurifn/ics-calendar-tidbyt-lambda/calendar"
	t "github.com/quesurifn/ics-calendar-tidbyt-lambda/types"
	"go.uber.org/zap"
)

func HandleRequest(ctx context.Context, event *t.IcsRequest) (*t.BaseResponse[t.IcsResponse], error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	defer func() {
		err := logger.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			logger.Fatal(err.Error())
		}
	}()

	if event == nil {
		logger.Fatal("event is nil")
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
		return nil, err
	}

	events, err := cal.ParseCalendar(data, event.TZ)
	if err != nil {
		logger.Error("Error", zap.Any("err", err))
		return nil, err
	}

	nextEvent := cal.NextEvent(events)
	if err != nil {
		logger.Error("Error", zap.Any("err", err))
		return nil, err
	}

	response := t.BaseResponse[t.IcsResponse]{
		Data: t.IcsResponse{
			EventName:         nextEvent.Name,
			EventStartTime:    nextEvent.StartTime,
			EventEndTime:      nextEvent.EndTime,
			EventLocation:     nextEvent.Location,
			TenMinuteWarning:  nextEvent.TenMinuteWarning,
			FiveMinuteWarning: nextEvent.FiveMinuteWarning,
			OneMinuteWarning:  nextEvent.OneMinuteWarning,
			InProgress:        nextEvent.InProgress,
		},
		Message: "Success",
	}

	return &response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
