build:
	docker build . -t aint/gs

test:
	go test -v ./...

run:
	docker-compose up gs-app

down:
	docker-compose down -v

integration:
	docker-compose up --build --abort-on-container-exit --exit-code-from test-runner
