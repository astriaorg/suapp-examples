set dotenv-filename := "examples/astria-passthrough/example.env"

build:
    forge build -C examples/astria-passthrough

run:
    echo $KETTLE_RPC
    go run examples/astria-passthrough/main.go
