run:
	docker-compose up --build

test:
	go test -v -race -cover ./...

build:
	docker build -t rate-limiter .

clean:
	docker-compose down -v
	docker ps -q | xargs -r docker stop
	docker system prune -af