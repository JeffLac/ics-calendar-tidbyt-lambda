TIDY ICS NEXT EVENT


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
        "eventName": "HARVEST",
        "eventStart": 1707498000,
        "eventEnd": 1707501600,
        "eventLocation": "https://us06web.zoom.us/j/xxx
        "tenMinuteWarning": false,
        "fiveMinuteWarning": false,
        "oneMinuteWarning": false,
        "inProgress": false
    },
    "message": null
}
```

