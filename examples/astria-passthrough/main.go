package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"

	"github.com/flashbots/suapp-examples/framework"
)

type config struct {
	ComposerURL string `env:"COMPOSER_URL, default=local"`
}

type bundle struct {
	Id      string        `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func bundleHandler(w http.ResponseWriter, r *http.Request) {
	kettleSignature := r.Header.Get("X-Flashbots-Signature")
	if kettleSignature == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "missing X-Flashbots-Signature header")
		return
	}
	fmt.Printf("Kettle signature: %s\n", kettleSignature)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "failed to read body")
		return
	}
	bundle := bundle{}
	err = json.Unmarshal(body, &bundle)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "failed to unmarshal bundle")
		return
	}
	fmt.Printf("Received bundle: %s\n", bundle)

	w.WriteHeader(http.StatusOK)
}

func main() {
	fr := framework.New()

	// set up server to receive post request
	go func() {
		http.HandleFunc("/rollupBundle", bundleHandler)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// create & fund suave account
	suaveAccount := framework.GeneratePrivKey()
	log.Printf("Suave account: %s", suaveAccount.Address().Hex())

	fundBalance := big.NewInt(100000000000000000)
	if err := fr.Suave.FundAccount(suaveAccount.Address(), fundBalance); err != nil {
		log.Fatal(err)
	}

	suaveBalance, err := fr.Suave.RPC().BalanceAt(context.Background(), suaveAccount.Address(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s funded with %s", suaveAccount.Address().Hex(), suaveBalance.String())

	// deploy suapp and load abi
	// suappHandle := fr.Suave.DeployContract("pass-through.sol/Passthrough.json")
	// passthrough, err := framework.ReadArtifact("passthrough.sol/Passthrough.json")
	// if err != nil {
	// 	panic(err)
	// }

	// craft "rollup tx" - for the same of the example this is just bytes
	// in reality, this would be a tx that the rollup node received from a user
	// rollupTx := []byte("hello, world!")

	// create ccr

	// submit ccr
	// suappHandle.SendTransaction()

	// wait for ccr to be included in a block

	//
}
