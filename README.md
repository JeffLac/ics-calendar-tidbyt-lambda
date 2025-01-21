#Tidbyt ICS Next Event


This is a server designed for tidyt app in development that displays the next event from ICS url.


E.g.

**POST lambda**
```


{
    "icsUrl": "https://outlook.office365.com/owa/calendar/xxx.com/xxx/calendar.ics",
    "tz": "America/Chicago"
}
```

**Response**

```
{
    "data": {
        "name": "HARVEST",
        "start": 1707498000,
        "end": 1707501600,
        "location": "https://us06web.zoom.us/j/xxx
        "detail": {
            "isToday": true,
            "isTomorrow": false,
            "isThisWeek": true,
            "minutesUntilStart": -736,
            "minutesUntilEnd": 703,
            "hoursToEnd": 11,
            "inProgress": true  
        }
        "tenMinuteWarning": false,
        "fiveMinuteWarning": false,
        "oneMinuteWarning": false,
        "inProgress": false
    },
    "message": null
}
```

