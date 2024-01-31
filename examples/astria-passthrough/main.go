package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flashbots/suapp-examples/framework"
)

type config struct {
	ComposerURL string `env:"COMPOSER_URL, default=local"`
}

type bundle struct {
	Id      string `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []byte `json:"params"`
}

func bundleHandler(w http.ResponseWriter, r *http.Request) {
	kettleSignature := r.Header.Get("X-Flashbots-Signature")
	if kettleSignature == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "missing X-Flashbots-Signature header")
		return
	}
	log.Printf("Kettle signature: %s\n", kettleSignature)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "failed to read body")
		log.Printf("Failed to read body: %s\n", err.Error())
		return
	}
	bundle := bundle{}
	err = json.Unmarshal(body, &bundle)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "failed to unmarshal bundle")
		log.Printf("Failed to unmarshal bundle: %s\n", err.Error())
		return
	}
	log.Printf("Received bundle: %s\n", bundle)

	w.WriteHeader(http.StatusOK)
}

func main() {
	fr := framework.New()

	// set up server to receive post request
	go func() {
		http.HandleFunc("/rollupBundle", bundleHandler)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// if private key env var not set create & fund suave account
	suaveAccount := &framework.PrivKey{}
	if os.Getenv("SUAVE_PRIVATE_KEY") == "" {
		log.Println("No private key provided, creating new account")
		suaveAccount = framework.GeneratePrivKey()
		log.Printf("Suave account: %s", suaveAccount.Address().Hex())
		log.Println(hex.EncodeToString(suaveAccount.MarshalPrivKey()))
		fundBalance := big.NewInt(100000000000000000)
		if err := fr.Suave.FundAccount(suaveAccount.Address(), fundBalance); err != nil {
			log.Fatal(err)
		}

	} else {
		// otherwise use the provided private key
		log.Println("Using provided private key")
		suaveAccount = framework.NewPrivKeyFromHex(os.Getenv("SUAVE_PRIVATE_KEY"))
		log.Printf("Suave account: %s", suaveAccount.Address().Hex())
	}

	suaveBalance, err := fr.Suave.RPC().BalanceAt(context.Background(), suaveAccount.Address(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s funded with %s", suaveAccount.Address().Hex(), suaveBalance.String())

	// get contract handle and abi
	suappHandle := &framework.Contract{}
	passthrough, err := framework.ReadArtifact("passthrough.sol/Passthrough.json")
	if err != nil {
		panic(err)
	}
	if os.Getenv("PASSTHROUGH_ADDR") == "" {
		log.Println("SUAPP address not provided, deploying passthrough contract")
		suappHandle = fr.Suave.DeployContract("passthrough.sol/Passthrough.json")
	} else {
		log.Println("Reading the provided SUAPP address")
		addr := common.Address{}
		err = addr.UnmarshalText([]byte(os.Getenv("PASSTHROUGH_ADDR")))
		if err != nil {
			log.Fatal(err)
		}
		suappHandle = fr.Suave.GetContract(addr, passthrough.Abi)
	}
	log.Printf("Using SUAPP address: %s", suappHandle.Address().Hex())

	log.Println("Creating rollup tx bytes")
	addr := suaveAccount.Address()
	// craft "rollup tx" - for the sake of the example this is just bytes
	// in reality, this would be a tx that the rollup node received from a user
	rollupTx, err := fr.L1.SignTx(suaveAccount, &types.LegacyTx{
		To:       &addr,
		Value:    big.NewInt(1234),
		Gas:      21000,
		GasPrice: big.NewInt(670189871),
	})
	if err != nil {
		log.Fatal(err)
	}

	rollupTxBytes, err := rollupTx.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	// submit tx with confidential inputs
	log.Println("Sending transaction to SUAPP")
	receipt := suappHandle.SendTransaction("makeBundle", []interface{}{}, rollupTxBytes)
	log.Printf("Transaction receipt: %s", receipt.TxHash.Hex())

	select {}
}
