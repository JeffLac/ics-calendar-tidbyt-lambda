package types

type Event struct {
	Name              string
	StartTime         int64
	EndTime           int64
	Location          *string
	TenMinuteWarning  bool
	FiveMinuteWarning bool
	OneMinuteWarning  bool
	InProgress        bool
}

type BaseResponse[t any] struct {
	Data    t      `json:"data"`
	Message string `json:"message"`
}

type IcsRequest struct {
	ICSUrl         string `json:"icsUrl"`
	ShowInProgress bool   `json:"showInProgress"`
	TZ             string `json:"tz"`
}

type IcsResponse struct {
	EventName         string  `json:"eventName"`
	EventStartTime    int64   `json:"eventStart"`
	EventEndTime      int64   `json:"eventEnd"`
	EventLocation     *string `json:"eventLocation"`
	TenMinuteWarning  bool    `json:"tenMinuteWarning"`
	FiveMinuteWarning bool    `json:"fiveMinuteWarning"`
	OneMinuteWarning  bool    `json:"oneMinuteWarning"`
	InProgress        bool    `json:"inProgress"`
}
