/* btckeygenie v1.0.0
 * https://github.com/vsergeev/btckeygenie
 * License: MIT
 */

package btckey

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"
)

/******************************************************************************/
/* Helper Functions */
/******************************************************************************/

func hex2bytes(hexstring string) (b []byte) {
	b, _ = hex.DecodeString(hexstring)
	return b
}

/******************************************************************************/
/* Base-58 Encode/Decode */
/******************************************************************************/

func TestBase58(t *testing.T) {
	var b58Vectors = []struct {
		bytes   []byte
		encoded string
	}{
		{hex2bytes("4e19"), "6wi"},
		{hex2bytes("3ab7"), "5UA"},
		{hex2bytes("ae0ddc9b"), "5T3W5p"},
		{hex2bytes("65e0b4c9"), "3c3E6L"},
		{hex2bytes("25793686e9f25b6b"), "7GYJp3ZThFG"},
		{hex2bytes("94b9ac084a0d65f5"), "RspedB5CMo2"},
	}

	/* Test base-58 encoding */
	for i := 0; i < len(b58Vectors); i++ {
		got := b58encode(b58Vectors[i].bytes)
		if got != b58Vectors[i].encoded {
			t.Fatalf("b58encode(%v): got %s, expected %s", b58Vectors[i].bytes, got, b58Vectors[i].encoded)
		}
	}
	t.Log("success b58encode() on valid vectors")

	/* Test base-58 decoding */
	for i := 0; i < len(b58Vectors); i++ {
		got, err := b58decode(b58Vectors[i].encoded)
		if err != nil {
			t.Fatalf("b58decode(%s): got error %v, expected %v", b58Vectors[i].encoded, err, b58Vectors[i].bytes)
		}
		if bytes.Compare(got, b58Vectors[i].bytes) != 0 {
			t.Fatalf("b58decode(%s): got %v, expected %v", b58Vectors[i].encoded, got, b58Vectors[i].bytes)
		}
	}
	t.Log("success b58decode() on valid vectors")

	/* Test base-58 decoding of invalid strings */
	b58InvalidVectors := []string{
		"5T3IW5p", // Invalid character I
		"6Owi",    // Invalid character O
	}

	for i := 0; i < len(b58InvalidVectors); i++ {
		got, err := b58decode(b58InvalidVectors[i])
		if err == nil {
			t.Fatalf("b58decode(%s): got %v, expected error", b58InvalidVectors[i], got)
		}
		t.Logf("b58decode(%s): got expected err %v", b58InvalidVectors[i], err)
	}
	t.Log("success b58decode() on invalid vectors")
}

/******************************************************************************/
/* Base-58 Check Encode/Decode */
/******************************************************************************/

func TestBase58Check(t *testing.T) {
	var b58CheckVectors = []struct {
		ver     uint8
		bytes   []byte
		encoded string
	}{
		{0x00, hex2bytes("010966776006953D5567439E5E39F86A0D273BEE"), "16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM"},
		{0x00, hex2bytes("000000006006953D5567439E5E39F86A0D273BEE"), "111112LbMksD9tCRVsyW67atmDssDkHHG"},
		{0x80, hex2bytes("0C28FCA386C7A227600B2FE50B7CAE11EC86D3BF1FBE471BE89827E19D72AA1D"), "5HueCGU8rMjxEXxiPuD5BDku4MkFqeZyd4dZ1jvhTVqvbTLvyTJ"},
	}

	/* Test base-58 check encoding */
	for i := 0; i < len(b58CheckVectors); i++ {
		got := b58checkencode(b58CheckVectors[i].ver, b58CheckVectors[i].bytes)
		if got != b58CheckVectors[i].encoded {
			t.Fatalf("b58checkencode(0x%02x, %v): got %s, expected %s", b58CheckVectors[i].ver, b58CheckVectors[i].bytes, got, b58CheckVectors[i].encoded)
		}
	}
	t.Log("success b58checkencode() on valid vectors")

	/* Test base-58 check decoding */
	for i := 0; i < len(b58CheckVectors); i++ {
		ver, got, err := b58checkdecode(b58CheckVectors[i].encoded)
		if err != nil {
			t.Fatalf("b58checkdecode(%s): got error %v, expected ver %v, bytes %v", b58CheckVectors[i].encoded, err, b58CheckVectors[i].ver, b58CheckVectors[i].bytes)
		}
		if ver != b58CheckVectors[i].ver || bytes.Compare(got, b58CheckVectors[i].bytes) != 0 {
			t.Fatalf("b58checkdecode(%s): got ver %v, bytes %v, expected ver %v, bytes %v", b58CheckVectors[i].encoded, ver, got, b58CheckVectors[i].ver, b58CheckVectors[i].bytes)
		}
	}
	t.Log("success b58checkdecode() on valid vectors")

	/* Test base-58 check decoding of invalid strings */
	b58CheckInvalidVectors := []string{
		"5T3IW5p", // Invalid base58
		"6wi",     // Missing checksum
		"6UwLL9Risc3QfPqBUvKofHmBQ7wMtjzm", // Invalid checksum
	}

	for i := 0; i < len(b58CheckInvalidVectors); i++ {
		ver, got, err := b58checkdecode(b58CheckInvalidVectors[i])
		if err == nil {
			t.Fatalf("b58checkdecode(%s): got ver %v, bytes %v, expected error", b58CheckInvalidVectors[i], ver, got)
		}
		t.Logf("b58checkdecode(%s): got expected err %v", b58CheckInvalidVectors[i], err)
	}
	t.Log("success b58checkdecode() on invalid vectors")
}

/******************************************************************************/
/* Common Key Pair Test Vectors */
/******************************************************************************/

var keyPairVectors = []struct {
	wif                    string
	wifc                   string
	priv_bytes             []byte
	address_compressed     string
	address_uncompressed   string
	pub_bytes_compressed   []byte
	pub_bytes_uncompressed []byte
	D                      *big.Int
	X                      *big.Int
	Y                      *big.Int
}{
	{
		"5J1F7GHadZG3sCCKHCwg8Jvys9xUbFsjLnGec4H125Ny1V9nR6V",
		"Kx45GeUBSMPReYQwgXiKhG9FzNXrnCeutJp4yjTd5kKxCitadm3C",
		hex2bytes("18E14A7B6A307F426A94F8114701E7C8E774E7F9A47E2C2035DB29A206321725"),
		"1PMycacnJaSqwwJqjawXBErnLsZ7RkXUAs",
		"16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM",
		hex2bytes("0250863AD64A87AE8A2FE83C1AF1A8403CB53F53E486D8511DAD8A04887E5B2352"),
		hex2bytes("0450863AD64A87AE8A2FE83C1AF1A8403CB53F53E486D8511DAD8A04887E5B23522CD470243453A299FA9E77237716103ABC11A1DF38855ED6F2EE187E9C582BA6"),
		hex2int("18E14A7B6A307F426A94F8114701E7C8E774E7F9A47E2C2035DB29A206321725"),
		hex2int("50863AD64A87AE8A2FE83C1AF1A8403CB53F53E486D8511DAD8A04887E5B2352"),
		hex2int("2CD470243453A299FA9E77237716103ABC11A1DF38855ED6F2EE187E9C582BA6"),
	},
	{
		"5JbDYniwPgAn3YqPUkVvrCQdJsjjFx2rV2EYeg5CAH3wNncziMm",
		"Kze2PJp755t9pFWaDUzgg9MHFtwbWyuBQgSnhHTWFwqy14NafA1S",
		hex2bytes("660527765029F5F1BC6DFD5821A7FF336C10EDA391E19BB4517DB4E23E5B112F"),
		"1ChaLikBC5E2uTCA7GZh9vaMQMuRt7h1yq",
		"17FBpEDgirwQJTvHT6ZgSirWSCbdTB9f76",
		hex2bytes("03A83B8DE893467D3A88D959C0EB4032D9CE3BF80F175D4D9E75892A3EBB8AB7E5"),
		hex2bytes("04A83B8DE893467D3A88D959C0EB4032D9CE3BF80F175D4D9E75892A3EBB8AB7E5370F723328C24B7A97FE34063BA68F253FB08F8645D7C8B9A4FF98E3C29E7F0D"),
		hex2int("660527765029F5F1BC6DFD5821A7FF336C10EDA391E19BB4517DB4E23E5B112F"),
		hex2int("A83B8DE893467D3A88D959C0EB4032D9CE3BF80F175D4D9E75892A3EBB8AB7E5"),
		hex2int("370F723328C24B7A97FE34063BA68F253FB08F8645D7C8B9A4FF98E3C29E7F0D"),
	},
	{
		"5KPaskZdrcPmrH3AFdpMF7FFBcYigwdrEfpBN9K5Ch4Ch6Bort4",
		"L4AgX1H3fyDWxVnXqbVzMGZsbqu11J9eKLKEkKgmYbo8bbs4K9Sq",
		hex2bytes("CF4DBE1ABCB061DB64CC87404AB736B6A56E8CDD40E9846144582240C5366758"),
		"1DP4edYeSPAF5UkXomAFKhsXwKq59r26aY",
		"1K1EJ6Zob7mr6Wye9mF1pVaU4tpDhrYMKJ",
		hex2bytes("03F680556678E25084A82FA39E1B1DFD0944F7E69FDDAA4E03CE934BD6B291DCA0"),
		hex2bytes("04F680556678E25084A82FA39E1B1DFD0944F7E69FDDAA4E03CE934BD6B291DCA052C10B721D34447E173721FB0151C68DE1106BADB089FB661523B8302A9097F5"),
		hex2int("CF4DBE1ABCB061DB64CC87404AB736B6A56E8CDD40E9846144582240C5366758"),
		hex2int("F680556678E25084A82FA39E1B1DFD0944F7E69FDDAA4E03CE934BD6B291DCA0"),
		hex2int("52C10B721D34447E173721FB0151C68DE1106BADB089FB661523B8302A9097F5"),
	},
	{
		"5KTzSQJFWc48YdgxXJPb7BhnHu98TUd6C8CDNw6D2dq8fVfC5G8",
		"L4W8X7q93fipJcwN4jkhYJEzub8survHv6kVdojz6DZSpYUmJYkM",
		hex2bytes("D94F024E82D787FB38369BEA7478AA61308DC2F7080ADDF69919A881490CFF48"),
		"1Hicf8AisGTeFqhNuSTw5m5UsYbHRxDxfj",
		"1CqhvePnxy5ZdvuunhZ7KzaqJVrNfXAk5E",
		hex2bytes("02692B035A2BB89C503E68A732596491524808BC2CC6A95061CD2CDE5151B34CD8"),
		hex2bytes("04692B035A2BB89C503E68A732596491524808BC2CC6A95061CD2CDE5151B34CD8B9FBFB401C7BDA0C77F161ADE0AA54688412E591DAF2E3A652DB00A533645B24"),
		hex2int("D94F024E82D787FB38369BEA7478AA61308DC2F7080ADDF69919A881490CFF48"),
		hex2int("692B035A2BB89C503E68A732596491524808BC2CC6A95061CD2CDE5151B34CD8"),
		hex2int("B9FBFB401C7BDA0C77F161ADE0AA54688412E591DAF2E3A652DB00A533645B24"),
	},
	{
		"5HpHagT65TZzG1PH3CSu63k8DbpvD8s5ip4nEB3kEsrgA9tXshp",
		"KwDiBf89QgGbjEhKnhXJuH7LrciVrZi3qYjgd9M7rFUFqJ5Vvujp",
		hex2bytes("0000000000000000000000000000000000000000000000000000000000000400"),
		"1G73bvYR97QGVb8bfeX2TqvSKietBDybQC",
		"17imJe7o4mpq2MMfZ328evDJQfbt6ShvxA",
		hex2bytes("03241FEBB8E23CBD77D664A18F66AD6240AAEC6ECDC813B088D5B901B2E285131F"),
		hex2bytes("04241FEBB8E23CBD77D664A18F66AD6240AAEC6ECDC813B088D5B901B2E285131F513378D9FF94F8D3D6C420BD13981DF8CD50FD0FBD0CB5AFABB3E66F2750026D"),
		hex2int("0000000000000000000000000000000000000000000000000000000000000400"),
		hex2int("241FEBB8E23CBD77D664A18F66AD6240AAEC6ECDC813B088D5B901B2E285131F"),
		hex2int("513378D9FF94F8D3D6C420BD13981DF8CD50FD0FBD0CB5AFABB3E66F2750026D"),
	},
}

var invalidPublicKeyBytesVectors = [][]byte{
	hex2bytes("0250863AD64A87AE8A2FE83C1AF1A8403CB53F53E486D8511DAD8A04887E5B23"),                                                                 /* Short compressed */
	hex2bytes("0450863AD64A87AE8A2FE83C1AF1A8403CB53F53E486D8511DAD8A04887E5B23522CD470243453A299FA9E77237716103ABC11A1DF38855ED6F2EE187E9C582B"), /* Short uncompressed */
	hex2bytes("03A83B8DFF93467D3A88D959C0EB4032FFFF3BF80F175D4D9E75892A3EBB8FF7E5"),                                                               /* Invalid compressed */
	hex2bytes("02c8e337cee51ae9af3c0ef923705a0cb1b76f7e8463b3d3060a1c8d795f9630fd"),                                                               /* Invalid compressed */
}

var wifInvalidVectors = []string{
	"5T3IW5p", // Invalid base58
	"6wi",     // Missing checksum
	"6Mcb23muAxyXaSMhmB6B1mqkvLdWhtuFZmnZsxDczHRraMcNG",    // Invalid checksum
	"huzKTSifqNioknFPsoA7uc359rRHJQHRg42uiKn6P8Rnv5qxV5",   // Invalid version byte
	"yPoVP5njSzmEVK4VJGRWWAwqnwCyLPRcMm5XyrKgYUpeXtGyM",    // Invalid private key byte length
	"Kx45GeUBSMPReYQwgXiKhG9FzNXrnCeutJp4yjTd5kKxCj6CAKu3", // Invalid private key suffix byte
}

func TestCheckWIF(t *testing.T) {
	/* Check valid vectors */
	for i := 0; i < len(keyPairVectors); i++ {
		got, err := CheckWIF(keyPairVectors[i].wif)
		if got == false {
			t.Fatalf("CheckWIF(%s): got false, error %v, expected true", keyPairVectors[i].wif, err)
		}
		got, err = CheckWIF(keyPairVectors[i].wifc)
		if got == false {
			t.Fatalf("CheckWIF(%s): got false, error %v, expected true", keyPairVectors[i].wifc, err)
		}
	}
	t.Log("success CheckWIF() on valid vectors")

	/* Check invalid vectors */
	for i := 0; i < len(wifInvalidVectors); i++ {
		got, err := CheckWIF(wifInvalidVectors[i])
		if got == true {
			t.Fatalf("CheckWIF(%s): got true, expected false", wifInvalidVectors[i])
		}
		t.Logf("CheckWIF(%s): got false, err %v", wifInvalidVectors[i], err)
	}
	t.Log("success CheckWIF() on invalid vectors")
}

func TestPrivateKeyDerive(t *testing.T) {
	var priv PrivateKey

	for i := 0; i < len(keyPairVectors); i++ {
		priv.D = keyPairVectors[i].D

		/* Derive public key from private key */
		priv.derive()

		if priv.X.Cmp(keyPairVectors[i].X) != 0 || priv.Y.Cmp(keyPairVectors[i].Y) != 0 {
			t.Fatalf("derived public key does not match test vector on index %d", i)
		}
	}
	t.Log("success PrivateKey derive()")
}

/******************************************************************************/
/* Bitcoin Private Key Import/Export */
/******************************************************************************/

func TestPrivateKeyFromBytes(t *testing.T) {
	var priv PrivateKey

	/* Check valid vectors */
	for i := 0; i < len(keyPairVectors); i++ {
		err := priv.FromBytes(keyPairVectors[i].priv_bytes)
		if err != nil {
			t.Fatalf("priv.FromBytes(D): got error %v, expected success on index %d", err, i)
		}
		if priv.D.Cmp(keyPairVectors[i].D) != 0 {
			t.Fatalf("private key does not match test vector on index %d", i)
		}
		if priv.X.Cmp(keyPairVectors[i].X) != 0 || priv.Y.Cmp(keyPairVectors[i].Y) != 0 {
			t.Fatalf("public key does not match test vector on index %d", i)
		}
	}
	t.Log("success PrivateKey FromBytes() on valid vectors")

	/* Invalid short private key */
	err := priv.FromBytes(keyPairVectors[0].priv_bytes[0:31])
	if err == nil {
		t.Fatalf("priv.FromBytes(D): got success, expected error")
	}
	/* Invalid long private key */
	err = priv.FromBytes(append(keyPairVectors[0].priv_bytes, []byte{0x00}...))
	if err == nil {
		t.Fatalf("priv.FromBytes(D): got success, expected error")
	}

	t.Log("success PrivateKey FromBytes() on invaild vectors")
}

func TestPrivateKeyToBytes(t *testing.T) {
	var priv PrivateKey

	/* Check valid vectors */
	for i := 0; i < len(keyPairVectors); i++ {
		priv.D = keyPairVectors[i].D
		b := priv.ToBytes()
		if bytes.Compare(keyPairVectors[i].priv_bytes, b) != 0 {
			t.Fatalf("private key bytes do not match test vector in index %d", i)
		}
	}
	t.Log("success PrivateKey ToBytes()")
}

func TestPrivateKeyFromWIF(t *testing.T) {
	var priv PrivateKey

	/* Check valid vectors */
	for i := 0; i < len(keyPairVectors); i++ {
		/* Import WIF */
		err := priv.FromWIF(keyPairVectors[i].wif)
		if err != nil {
			t.Fatalf("priv.FromWIF(%s): got error %v, expected success", keyPairVectors[i].wif, err)
		}
		if priv.D.Cmp(keyPairVectors[i].D) != 0 {
			t.Fatalf("private key does not match test vector on index %d", i)
		}
		if priv.X.Cmp(keyPairVectors[i].X) != 0 || priv.Y.Cmp(keyPairVectors[i].Y) != 0 {
			t.Fatalf("public key does not match test vector on index %d", i)
		}

		/* Import WIFC */
		err = priv.FromWIF(keyPairVectors[i].wifc)
		if err != nil {
			t.Fatalf("priv.FromWIF(%s): got error %v, expected success", keyPairVectors[i].wifc, err)
		}
		if priv.D.Cmp(keyPairVectors[i].D) != 0 {
			t.Fatalf("private key does not match test vector on index %d", i)
		}
		if priv.X.Cmp(keyPairVectors[i].X) != 0 || priv.Y.Cmp(keyPairVectors[i].Y) != 0 {
			t.Fatalf("public key does not match test vector on index %d", i)
		}
	}
	t.Log("success PrivateKey FromWIF() on valid vectors")

	/* Check invalid vectors */
	for i := 0; i < len(wifInvalidVectors); i++ {
		err := priv.FromWIF(wifInvalidVectors[i])
		if err == nil {
			t.Fatalf("priv.FromWIF(%s): got success, expected error", wifInvalidVectors[i])
		}
		t.Logf("priv.FromWIF(%s): got err %v", wifInvalidVectors[i], err)
	}
	t.Log("success PrivateKey FromWIF() on invalid vectors")
}

func TestPrivateKeyToWIF(t *testing.T) {
	var priv PrivateKey

	/* Check valid vectors */
	for i := 0; i < len(keyPairVectors); i++ {
		/* Export WIF */
		priv.D = keyPairVectors[i].D
		wif := priv.ToWIF()
		if wif != keyPairVectors[i].wif {
			t.Fatalf("priv.ToWIF() %s != expected %s", wif, keyPairVectors[i].wif)
		}
	}
	t.Log("success PrivateKey ToWIF()")

	/* Check valid vectors */
	for i := 0; i < len(keyPairVectors); i++ {
		/* Export WIFC */
		priv.D = keyPairVectors[i].D
		wifc := priv.ToWIFC()
		if wifc != keyPairVectors[i].wifc {
			t.Fatalf("priv.ToWIFC() %s != expected %s", wifc, keyPairVectors[i].wifc)
		}
	}
	t.Log("success PrivateKey ToWIFC()")
}

/******************************************************************************/
/* Bitcoin Public Key Import/Export */
/******************************************************************************/

func TestPublicKeyToBytes(t *testing.T) {
	var pub PublicKey

	/* Check valid vectors */
	for i := 0; i < len(keyPairVectors); i++ {
		pub.X = keyPairVectors[i].X
		pub.Y = keyPairVectors[i].Y

		bytes_compressed := pub.ToBytes()
		if bytes.Compare(keyPairVectors[i].pub_bytes_compressed, bytes_compressed) != 0 {
			t.Fatalf("public key compressed bytes do not match test vectors on index %d", i)
		}

		bytes_uncompressed := pub.ToBytesUncompressed()
		if bytes.Compare(keyPairVectors[i].pub_bytes_uncompressed, bytes_uncompressed) != 0 {
			t.Fatalf("public key uncompressed bytes do not match test vectors on index %d", i)
		}
	}
	t.Log("success PublicKey ToBytes() and ToBytesUncompressed()")
}

func TestPublicKeyFromBytes(t *testing.T) {
	var pub PublicKey

	/* Check valid vectors */
	for i := 0; i < len(keyPairVectors); i++ {
		err := pub.FromBytes(keyPairVectors[i].pub_bytes_compressed)
		if err != nil {
			t.Fatalf("pub.FromBytes(): got error %v, expected success on index %d", err, i)
		}
		if pub.X.Cmp(keyPairVectors[i].X) != 0 || pub.Y.Cmp(keyPairVectors[i].Y) != 0 {
			t.Fatalf("public key does not match test vectors on index %d", i)
		}

		err = pub.FromBytes(keyPairVectors[i].pub_bytes_uncompressed)
		if err != nil {
			t.Fatalf("pub.FromBytes(): got error %v, expected success on index %d", err, i)
		}
		if pub.X.Cmp(keyPairVectors[i].X) != 0 || pub.Y.Cmp(keyPairVectors[i].Y) != 0 {
			t.Fatalf("public key does not match test vectors on index %d", i)
		}
	}
	t.Log("success PublicKey FromBytes() on valid vectors")

	/* Check invalid vectors */
	for i := 0; i < len(invalidPublicKeyBytesVectors); i++ {
		err := pub.FromBytes(invalidPublicKeyBytesVectors[i])
		if err == nil {
			t.Fatal("pub.FromBytes(): got success, expected error")
		}
		t.Logf("pub.FromBytes(): got error %v", err)
	}
	t.Log("success PublicKey FromBytes() on invalid vectors")
}

func TestToAddress(t *testing.T) {
	var pub PublicKey

	/* Check valid vectors */
	for i := 0; i < len(keyPairVectors); i++ {
		pub.X = keyPairVectors[i].X
		pub.Y = keyPairVectors[i].Y

		address_compressed := pub.ToAddress()
		if address_compressed != keyPairVectors[i].address_compressed {
			t.Fatalf("public key compressed address does not match test vectors on index %d", i)
		}

		address_uncompressed := pub.ToAddressUncompressed()
		if address_uncompressed != keyPairVectors[i].address_uncompressed {
			t.Fatalf("public key uncompressed address does not match test vectors on index %d", i)
		}
	}
	t.Log("success PublicKey ToAddress() and ToAddressUncompressed()")
}

/******************************************************************************/
/* Generating Keys */
/******************************************************************************/

func TestGenerateKey(t *testing.T) {
	/* Generate a keypair and check public key validity */
	priv1, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey(): got error %v", err)
	}
	if !secp256k1.IsOnCurve(priv1.PublicKey.Point) {
		t.Fatalf("GenerateKey(): public key not on curve")
	}

	/* Generate another keypair and check public key validity */
	priv2, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey(): got error %v", err)
	}
	if !secp256k1.IsOnCurve(priv1.PublicKey.Point) {
		t.Fatalf("GenerateKey(): public key not on curve")
	}

	/* Compare keypair private keys */
	if priv1.D == priv2.D {
		t.Fatalf("generated duplicate private keys")
	}
	/* Compare keypair public keys */
	if priv1.X.Cmp(priv2.X) == 0 && priv2.Y.Cmp(priv2.Y) == 0 {
		t.Fatalf("generated duplicate public keys")
	}

	t.Log("success GenerateKey()")
}
