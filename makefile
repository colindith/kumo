dev:
		go run httpd/main.go
build:
		go build -o ./app httpd/main.go
run:
		./app
worker:
		go run services/machinery/machinery.go -c services/machinery/config.yml worker