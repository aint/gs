# Overview
At first I'd like to apologize for a lot of typos and other mistakes. Unfortunately due to a lack of time I had no chance to review all the stuff.
Basically this app is just an abstraction on top of InfluxDB. IMO the best solution would be just to use InfluxDB directly without writing any code :smiley:

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
Actually InfluxDB supports such sufixes natively. But I implemented this feature by myself because I found it interesting. So there are probably some bugs :smiley:
# Timeline

## Thursday
Received and read the test assignment at the evening. At first I was like - ~~wat? I gon't get it~~ well, we need some DB, caching, REST API, docker and stuff. Seems pretty easy. What is great, that there is a place for experiments.

## Friday
So let's pick up something as a storage because it's the most challenging part. The `params` field seems fancy so maybe MongoDB? Nah, ~~I'm too old for this~~ MongoDb is a meme and nowadays even MySQL is able to store data in the JSON format. So let's take some old good RDBMS, add some indexes and that's it.

But why so serious. This a demo app so why not to jump in with some crazy idea? Hm, let's take Redis and store all the data solely in memory. Need performance? Get it! It will cause a hell lot of questions and objections. But why not? At least it will be funny.

But then I had a look at the description again. Well, we have events with a timestamp and data. Hm, looks like the perfect fit for time series DB! `event_type` is a tag, `params` is a value and of course `timestamp`. InfluxDB is written in Go, has outstanding performance, the connection is made via HTTP which is stateless. And I have some experience with this DB. So let's take InfluxDB as a storage.

## Sunday

It's time to code.

### Tech stack

- Go 1.12 with `mod` for dependency management
- InfluxDB as a storage
- Docker and Docker Compose to manage the orchestration
- Redis as a distrubuted cache (note: unfortunately it's not done due to a lack of time)
    - A local in-memory cache is not applicable for this app at all. It's worth to intoduce some distibuted cache but only for API that operates with absolute time intervals. It doesn't make a sense to cache relative parameters like the following `start=2h ago`
- REST API as a comminication interface
    - gRPC looks interesting and worth trying out but I have a limited time. So maybe next time.
    - GraphQL seems not good fit to me
    - API should supports processing new events in batch as we have to deal with thousands of events coming from a mobile SDKs every second.


### Project structure

I'm a person from the Java world which is notorious by overengineering and abusing of abstraction. So maybe you're expecting to see there some packages like `model`, `handler`, etc. No way. I like to follow the YAGNI principle. This is a small demo app that will contain a copule of files. So a single package can actually work very well here. Thus to avoid the issue with running like `go run main.go handler.go influxdb.go` I just used `app` package and the main file in the root directory.

### Testing

I advocate for BDD testing because tests that follow BDD principles are more expressive and readable.

# Improvments

- [ ] swagger
- [ ] structured logs
- [ ] caching
