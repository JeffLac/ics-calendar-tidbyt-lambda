package calendar

import (
	"sort"
	"strings"
	"time"

	"github.com/apognu/gocal"
	t "github.com/quesurifn/ics-calendar-tidbyt-lambda/ics/types"
	"go.uber.org/zap"
	"gopkg.in/resty.v1"
)

type Calendar struct {
	Logger *zap.Logger
	TZMap  map[string]string
}

func (c Calendar) DownloadCalendar(url string) (string, error) {
	client := resty.New()

	resp, err := client.R().Get(url)
	if err != nil {
		return "", err
	}

	return resp.String(), err
}

func (c Calendar) ParseCalendar(data string, tz string) ([]t.Event, error) {
	gocal.SetTZMapper(func(s string) (*time.Location, error) {
		override := ""
		if val, ok := c.TZMap[s]; ok {
			override = val
		}
		if override != "" {
			loc, err := time.LoadLocation(override)
			if err != nil {
				c.Logger.Error("Error", zap.Any("err", err))
				return nil, err
			}
			return loc, nil
		}

		loc, err := time.LoadLocation(s)
		if err != nil {
			c.Logger.Error("Error", zap.Any("err", err))
			return nil, err
		}
		return loc, nil
	})

	usersLoc, err := time.LoadLocation(tz)
	if err != nil {
		c.Logger.Error("Error", zap.Any("err", err))
		return nil, err
	}

	parser := gocal.NewParser(strings.NewReader(data))
	start, end := time.Now().In(usersLoc), time.Now().AddDate(0, 0, 7).In(usersLoc)
	parser.Start, parser.End = &start, &end

	parser.Parse()

	var events []t.Event
	for _, e := range parser.Events {
		events = append(events, t.Event{
			Name:      e.Summary,
			StartTime: e.Start.Unix(),
			EndTime:   e.End.Unix(),
			Location:  &e.Location,
		})
	}

	c.Logger.Info("ParseCalendar", zap.Any("events", events))

	return events, nil
}

func (c Calendar) NextEvent(events []t.Event, tz string) (*t.Event, error) {
	if len(events) == 0 {
		return nil, nil
	}

	location, err := time.LoadLocation(tz)
	if err != nil {
		c.Logger.Error("Error", zap.Any("err", err))
		return nil, err
	}

	now := time.Now().In(location).Unix()

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime < events[j].StartTime
	})
	next := events[0]

	next.Detail = &t.EventDetail{}
	next.Detail.InProgress = now >= next.StartTime
	next.Detail.IsThisWeek = now < next.StartTime+7*24*60*60
	next.Detail.IsToday = time.Unix(now, 0).Day() == time.Unix(next.StartTime, 0).Day()
	next.Detail.IsTomorrow = time.Unix(now, 0).Day() == time.Unix(next.StartTime, 0).Day()-1
	next.Detail.MinutesUntilStart = int(next.StartTime-now) / 60
	next.Detail.MinutesUntilEnd = int(next.EndTime-now) / 60
	next.Detail.HoursToEnd = int(next.EndTime-now) / 60 / 60

	c.Logger.Info("NextEvent", zap.Any("nextEvent", next))
	c.Logger.Info("Now", zap.Any("now", now))

	return &next, nil
}
