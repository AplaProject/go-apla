/* btckeygenie v1.0.0
 * https://github.com/vsergeev/btckeygenie
 * License: MIT
 */

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/vsergeev/btckeygenie/btckey"
	"log"
	"os"
)

func byteString(b []byte) (s string) {
	s = ""
	for i := 0; i < len(b); i++ {
		s += fmt.Sprintf("%02X", b[i])
	}
	return s
}

func main() {
	/* Redirect fatal errors to stderr */
	log.SetOutput(os.Stderr)

	var priv btckey.PrivateKey
	var err error

	/* Help/Usage */
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Printf("Usage: %s [WIF/WIFC]\n\n", os.Args[0])
		fmt.Println("btckeygenie v1.0.0 - https://github.com/vsergeev/btckeygenie")
		os.Exit(0)
	}

	/* Import WIF from first argument */
	if len(os.Args) > 1 {
		err = priv.FromWIF(os.Args[1])
		if err != nil {
			log.Fatalf("Importing WIF: %s\n", err)
		}
	} else {
		/* Generate a new Bitcoin keypair */
		priv, err = btckey.GenerateKey(rand.Reader)
		if err != nil {
			log.Fatalf("Generating keypair: %s\n", err)
		}
	}

	/* Convert to Compressed Address */
	address_compressed := priv.ToAddress()
	/* Convert to Public Key Compressed Bytes (33 bytes) */
	pub_bytes_compressed := priv.PublicKey.ToBytes()
	pub_bytes_compressed_str := byteString(pub_bytes_compressed)
	pub_bytes_compressed_b64 := base64.StdEncoding.EncodeToString(pub_bytes_compressed)

	/* Convert to Uncompressed Address */
	address_uncompressed := priv.ToAddressUncompressed()
	/* Convert to Public Key Uncompresed Bytes (65 bytes) */
	pub_bytes_uncompressed := priv.PublicKey.ToBytesUncompressed()
	pub_bytes_uncompressed_str := byteString(pub_bytes_uncompressed)
	pub_bytes_uncompressed_b64 := base64.StdEncoding.EncodeToString(pub_bytes_uncompressed)

	/* Convert to WIF and WIFC */
	wif := priv.ToWIF()
	wifc := priv.ToWIFC()
	/* Convert to Private Key Bytes (32 bytes) */
	pri_bytes := priv.ToBytes()
	pri_bytes_str := byteString(pri_bytes)
	pri_bytes_b64 := base64.StdEncoding.EncodeToString(pri_bytes)

	fmt.Printf("Bitcoin Address (Compressed)        %s\n", address_compressed)
	fmt.Printf("Public Key Bytes (Compressed)       %s\n", pub_bytes_compressed_str)
	fmt.Printf("Public Key Base64 (Compressed)      %s\n", pub_bytes_compressed_b64)
	fmt.Println()
	fmt.Printf("Bitcoin Address (Uncompressed)      %s\n", address_uncompressed)
	fmt.Printf("Public Key Bytes (Uncompressed)     %s\n", pub_bytes_uncompressed_str[0:65])
	fmt.Printf("                                    %s\n", pub_bytes_uncompressed_str[65:])
	fmt.Printf("Public Key Base64 (Uncompressed)    %s\n", pub_bytes_uncompressed_b64)
	fmt.Println()
	fmt.Printf("Private Key WIFC (Compressed)       %s\n", wifc)
	fmt.Printf("Private Key WIF (Uncompressed)      %s\n", wif)
	fmt.Printf("Private Key Bytes                   %s\n", pri_bytes_str)
	fmt.Printf("Private Key Base64                  %s\n", pri_bytes_b64)
}
