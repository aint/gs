[![Build Status](https://travis-ci.com/aint/gs.svg?branch=master)](https://travis-ci.com/aint/gs)

# Overview

Basically this app is just an abstraction on top of InfluxDB.

# Quick Start

To start the app just run the the following command

```
$ make run
```

The app should now be running at http://localhost:8080

Send some POST request with curl to save events
```
curl -X POST \
  http://localhost:8080/events \
  -H 'Content-Type: application/json' \
  -d '[
    {
        "event_type": "link_clicked",
        "ts": 1558892660,
        "params": {
            "url": "localhost:5000/app"
        }
    },
        {
        "event_type": "link_clicked",
        "ts": 1558892797,
        "params": {
            "url": "localhost:6000/app"
        }
    }
]'
```

and to get events
```
curl -X GET 'http://localhost:8080/events/relative?type=link_clicked&start=10d&end=1m'
```

You can use the following sufixes time filtering: `(m|minute|minutes|h|hour|hours|d|day|days|w|weeks|weeks)`. The `end` param is optional.

# Improvments

- [ ] swagger
- [ ] structured logs
- [ ] caching
- [ ] caching for TravisCI build
