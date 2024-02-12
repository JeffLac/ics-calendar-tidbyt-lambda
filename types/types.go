package types

type Event struct {
	Name       string
	StartTime  int64
	EndTime    int64
	Location   *string
	InProgress bool
}

type BaseResponse struct {
	Data  *IcsResponse   `json:"data"`
	Error *ErrorResponse `json:"message"`
}

type IcsRequest struct {
	ICSUrl         string `json:"icsUrl" validate:"required,url"`
	ShowInProgress bool   `json:"showInProgress" validate:"boolean"`
	TZ             string `json:"tz" validate:"required,timezone"`
}

type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type IcsResponse struct {
	EventName      string  `json:"eventName"`
	EventStartTime int64   `json:"eventStart"`
	EventEndTime   int64   `json:"eventEnd"`
	EventLocation  *string `json:"eventLocation"`
	InProgress     bool    `json:"inProgress"`
}
