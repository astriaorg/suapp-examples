set dotenv-filename := "examples/astria-passthrough/example.env"

build:
    forge build -C examples/astria-passthrough

rebuild:
    forge clean
    just build

run:
    clear
    go run examples/astria-passthrough/main.go

test rollupBundle:
    curl -X POST -H "X-Flashbots-Signature: 123" \
        -d '{"id": "1", "jsonrpc": "2.0", "method": "rollupBundle", "params": "dGVzdA=="}' \
        localhost:8080/rollupBundle

geth:
    suave-geth --suave.dev --suave.eth.external-whitelist localhost