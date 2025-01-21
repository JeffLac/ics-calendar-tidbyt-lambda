package calendar

import (
	"sort"
	"strings"
	"time"
	"fmt"

	"github.com/apognu/gocal"
	t "github.com/quesurifn/ics-calendar-tidbyt-lambda/types"
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

	//look back one day for all day events
	//this introduces additional complexity because now old events are showing up
	start, end := time.Now().AddDate(0, 0, -1).In(usersLoc), time.Now().AddDate(0, 0, 7).In(usersLoc)
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

func (c Calendar) NextEvent(events []t.Event, tz string) *t.Event {
	if len(events) == 0 {
		return nil
	}
	now := time.Now().Unix()

	offset, err := GetOffset(tz)
	if err != nil {
		fmt.Println("Error:", err)

		//set offset to UTC if error
		offset = 0
	}

	//loop through all events in slice
    for i := range events {
		//look for all day events
		if (isAllDayEvent(events[i])){
			//if event starts and ends at midnight UTC, then it is an all day event
			//convert start and end times from UTC to the correct timezone
			events[i].StartTime = events[i].StartTime - int64(offset)
			events[i].EndTime = events[i].EndTime - int64(offset)

		}		
    }

	//build a list of events that are in progress
	eventsInProgress := FilterInProgress(events)

    //if there are events in progress,  use the filtered list
	if len(eventsInProgress) > 0 {
		events = eventsInProgress
		//if there are multiple events in progress, sort by end time
		//this will put timed events that are in progress ahead of all day events
		sort.Slice(events, func(i, j int) bool {
			return events[i].EndTime < events[j].EndTime
		})
	}else{
		//there are no events in progress, so sort by events that are starting first
		//this will put all day events (starting at midnight) ahead of timed events
		sort.Slice(events, func(i, j int) bool {
			return events[i].StartTime < events[j].StartTime
		})
	}

	//assume we are going to use the first event as the next event
	next := events[0]

	//but sometimes events that have already ended get pulled because of the 1 day look back
	for i := 0; next.EndTime < now; i++ {
		next = events[i]
	}


	next.InProgress = now >= next.StartTime

	c.Logger.Info("NextEvent", zap.Any("nextEvent", next))
	c.Logger.Info("Now", zap.Any("now", now))

	return &next
}

//isAllDayEvent function verifies that both the Start is at midnight UTC 
//and End is at one second until midnight UTC.
//all day events can span multiple days, but they always start and end at midnight
func isAllDayEvent(e t.Event) bool {
	//look for divisible by 86400, so starts at UTC midnight
	startMidnightUTC := e.StartTime % 86400 == 0
	//look for to see if it ends at 11:59:59 so add a second and divide by 86400
	endMidnightUTC := (e.EndTime +1) % 86400 == 0	

	return startMidnightUTC && endMidnightUTC
}


//FilterInProgress filters events that are currently in progress
func FilterInProgress(events []t.Event) []t.Event {
	now := time.Now().Unix()
	// Slice to store events with InProgress = true
	var filteredEvents []t.Event

	// Iterate through the input slice
	for _, event := range events {
		//look to see if we are in between start time and end time
		if event.StartTime <= now && event.EndTime >= now {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// Return the filtered slice (empty if no matches)
	return filteredEvents
}


// GetOffset returns the offset in seconds for a given timezone
func GetOffset(timezone string) (int, error) {
	// Load the location
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, fmt.Errorf("failed to load location: %w", err)
	}

	// Get the current time in the specified location
	now := time.Now().In(loc)

	// Get the offset in seconds from UTC
	_, offset := now.Zone()

	return offset, nil
}
