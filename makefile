dev:
		go run httpd/main.go
build:
		go build -o ./app httpd/main.go
run:
		go build -o ./app httpd/main.go
		./app
worker:
		go run packages/machinery/example/machinery.go -c packages/machinery/example/config.yml worker