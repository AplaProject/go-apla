package parser

import ( "testing"
	"fmt"
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type InputTest struct {
	Head string  `json:"head"`
	Tx   string  `json:"tx"`
}

type BlockTest struct {
	Input  InputTest `json:"input"`
}

var (
	//       1     4       4       len       1
	//head = 0, block_id, time, wallet_id, cb_id
	//         1     4      len        len         len         len        len
	//tx   = type, time, wallet_id, citizen_id, public_key, node_public, host
	rules = []string{
		`1 4 4 s 1`,
		`1 4 s s h h s`,
	}
	private_key = `1f 2a b1 19 82 9f 34 82 ad e1 d4 fc 61 09 99 7f ac a9 76 98 e9 de 2d 58 00 a2 a9 eb a4 3b 56 3a`
	node_private_key = `f0 92 70 5b 6e 31 0f d3 6b 8c d4 eb 59 e7 80 57 4c 02 a5 d0 e7 18 1e ce 08 d9 98 c9 27 a1 4a 6b`
	input = []byte(`
[{ "input": {  
		"head" : "0 1 1472623022 1 0",
		"tx" : 	"1 1472623322 1 0 *ac 27 f8 0c 05 36 d0 34 dd 57 85 29 0f 40 11 65 57 1b 
		80 de 4a f0 ce 0b 13 33 f0 74 5f 5a e1 92 c4 83 01 f5 1b b6 8f cd 2b 13 
		8d c4 ad 3b 79 11 7b b5 89 45 ee c3 51 38 ec 8f 5f 4c 3e b4 ff 4f* 
		*6d c7 4e 28 cb 38 cc a5 13 52 b3 2b 1c a9 f5 86 58 a7 90 28 fe a9 ab 58 
		4b a4 49 c0 7a ae 01 fd 5a 2f 33 36 03 21 43 78 46 c3 4c d4 6e 94 e2 15 
		8b 71 e5 05 a1 61 02 00 43 fe 0b 79 a0 17 99 3e* 
		127.0.0.1"
	}
 }
]`)
    blocks []*BlockTest
)

func LoadData( input []byte, t *testing.T) []*BlockTest {
	// Delete \r\n\t inside strings
	prep := make([]byte, 0, len(input))
	quote := false
	for _,ch := range input {
		switch ch {
			case '"': quote = !quote
			case '\r', '\t', '\n': if quote {
				continue
			} 
		}
		prep = append(prep, ch)
	}
	ret := make([]*BlockTest,0)
	if err := json.Unmarshal( prep, &ret); err != nil {
		t.Error(`LoadData`, err)
	}
	return ret
	
}

func TestFirstBlock(t *testing.T) {
	fmt.Printf("%x %x %x\r\n",utils.EncodeLenInt64(1), utils.EncodeLenInt64(0), utils.EncodeLenInt64(65000))
	b127 := append( utils.EncodeLenInt64(0))//, 1, 2, 3, 4)
	fmt.Println( utils.DecodeLenInt64(&b127))
	fmt.Println(`b`, b127)
	blocks = LoadData(input, t) 
	t.Log(`First Block`)
	//t.Error(`Ooops`)
}

func TestMain(m *testing.M) {
//	m.Log(`Main`)
    fmt.Println(`OK`)
	m.Run()
}