package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
)

func main() {
	pkhex := os.Args[1]
	txFileName := os.Args[2]
	txData, err := ioutil.ReadFile(txFileName)
	if err != nil {
		log.Fatal("Reading tx file:", err)
	}
	pb, err := hex.DecodeString(string(pkhex))
	if err != nil {
		log.Fatal("Decoding pk from hex:", err)
	}

	rtx := &blockchain.Transaction{}
	if err = rtx.Unmarshal(txData); err != nil {
		log.Fatal("Unmarshalling tx data:", err)
	}

	hash, err := rtx.Hash()
	if err != nil {
		log.Fatal("hashing transaction:", err)
	}
	sign, err := crypto.Sign([]byte(pb), hash)
	if err != nil {
		log.Fatal("signing transaction:", err)
	}
	sign = converter.EncodeLengthPlusData(sign)
	dst := make([]byte, hex.EncodedLen(len(sign)))
	hex.Encode(dst, sign)
	fmt.Printf("%s\n", dst)
}
