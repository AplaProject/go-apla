package utils

import (
	"encoding/base64"
	"fmt"
)

func fillPct() [391]map[string]string {
	pctb64 := `AAAAAAE9AAJ6AAO2AATyAAYuAAdpAAikAAnfAAsZAAxTAA2NAA7HABAAABE5ABJxABOpABThABYZABdQABiHABm+ABr1ABwrAB1gAB6WAB/LACEAACI1ACNpACSdACXRACcEACg3AClqACqdACvPAC0BAC4yAC9kADCVADHGADL2ADQmADVWADaGADe1ADjkADoTADtBADxvAD2dAD7LAD/4AEElAEJSAEN+AESqAEXWAEcCAEgtAElYAEqDAEutAEzXAE4BAE8rAFBUAFF9AFKmAFPOAFT3AFYfAFdGAFhuAFmVAFq8AFviAF0IAF4uAF9UAGB6AGGfAGLEAGPoAGUNAGYxAGdVAGh4AGmcAGq/AGviAG0EAG4mAG9IAHBqAHGLAHKtAHPOAHTuAHYPAHcvAHhPAHluAHqOAHutAHzMAH3qAH8IAIAnAIFEAIJiAIN/AIScAIW5AIbWAIfyAIkOAIoqAItFAIxgAI17AI6WAI+xAJDLAJHlAJL+AJQYAJUxAJZKAJdjAJh7AJmUAJqsAJvDAJzbAJ3yAJ8JAKAgAKE3AKJNAKNjAKR5AKWOAKajAKe5AKjNAKniAKr2AKwKAK0eAK4yAK9FALBYALFrALJ+ALOQALSjALW0ALbGALfYALjpALn6ALsLALwbAL0sAL48AL9LAMBbAMFqAMJ6AMOIAMSXAMWmAMa0AMfCAMjQAMndAMrqAMv3AM0EAM4RAM8dANApANE1ANJBANNNANRYANVjANZuANd4ANiDANmNANqXANugANyqAN2zAN68AN/FAODNAOHWAOwdAPZPAQBsAQpzARRmAR5FASgQATHHATtrAUT7AU55AVfkAWE+AWqFAXO7AXzfAYXyAY70AZfmAaDHAamYAbJZAbsKAcOsAcw+AdTBAd02AeWcAe3zAfY8Af53AgakAg7EAhbVAh7aAibRAi67AjaYAj5pAkYtAk3lAlWQAl0wAmTDAmxLAnPHAns3AoKcAon2ApFFApiJAp/CAqbwAq4UArUtArw8AsNAAso7AtErAtgSAt7vAuXCAuyLAvNLAvoCAwCwAwdUAw3vAxSBAxsLAyGLAygDAy5yAzTZAzs3A0GNA0fbA04hA1ReA1qUA3kqA5cDA7QoA9CiA+x4BAeyBCJVBDxpBFX0BG76BIeBBJ+PBLcnBM5OBOUJBPtaBRFHBSbSBTv+BVDQBWVJBXltBY0/BaDBBbP1BcbeBdl/BevZBf3uBg/BBiFSBjKlBkO7BlSWBmU2BnWeBoXPBpXLBqWSBrUnBtO9BvGWBw68Bys2B0cMB2JFB3zpB5b9B7CHB8mNB+IVB/oiCBG6CCjiCD+cCFXuCGvaCIFlCJaSCKtjCL/dCNQBCOfSCPtUCQ6JCSFyCTQSCUZsCViBCWpUCXvmCY05CZ5PCa8pCb/KCdAyCeBjCfBeCgAmCg+7Ch8eCi5RCj1UCkwqClrTCmlPCnehCoXJCpPICqGfCq9PCrzZCso9Ctd8CuSYCvGQCv5mCwsbCxeuCyQhCzB0CzyoC0i+C1S2C2CQC2xOC3fvC4N1C47gC5ow`
	b64, _ := base64.StdEncoding.DecodeString(pctb64)
	data1 := BinToHex(b64)
	var arr [391]map[string]string
	j := 0
	for i := 0.0; i < 20; i = i + 0.1 {
		arr[j] = make(map[string]string)
		arr[j][Float64ToStr(i)] = pct_(&data1)
		j++
	}
	for i := 20; i < 100; i = i + 1 {
		arr[j] = make(map[string]string)
		arr[j][IntToStr(i)] = pct_(&data1)
		j++
	}
	for i := 100; i < 300; i = i + 5 {
		arr[j] = make(map[string]string)
		arr[j][IntToStr(i)] = pct_(&data1)
		j++
	}
	for i := 300; i <= 1000; i = i + 10 {
		arr[j] = make(map[string]string)
		arr[j][IntToStr(i)] = pct_(&data1)
		j++
	}
	return arr
}
func ArraySearch(value string, arr [391]map[string]string) string {
	for _, v := range arr {
		for k0, v0 := range v {
			if v0 == value {
				return k0
			}
		}
	}
	return ""
}
func CheckPct(pct float64) bool {
	arr := fillPct()
	for _, pct0 := range arr {
		for y, _ := range pct0 {
			if StrToFloat64(ClearNull(y, 2)) == pct {
				return true
			}
		}
	}
	return false
}

func CheckPct0(pct float64) bool {
	arr := fillPct()
	for _, pct0 := range arr {
		for _, sec := range pct0 {
			if sec == Float64ToStr(pct) {
				return true
			}
		}
	}
	return false
}
func pct_(data *[]byte) string {
	data_ := *data
	//fmt.Printf("data_=%s\n", data_)
	pct0 := fmt.Sprintf("0.000000%07d", HexToDec(string(data_[0:6])))
	//fmt.Println("pct0", pct0)
	//pct:="0.000000"+str_pad(hexdec(substr($data, 0, 6)), 7, "0", STR_PAD_LEFT);
	if len(data_) >= 6 {
		*data = data_[6:]
	}
	return pct0
}

// массив, в котором будет искаться максимальное кол-во голосов, должен быть стандартизирован
// входные данные уже были ранее проверены
func MakePctArray(pctArray map[string]int64) []map[int64]int64 {
	arr := fillPct()
	var i int64
	var result []map[int64]int64
	//log.Println("pctArray", pctArray)
	for _, pct := range arr {
		for _, sec := range pct {
			//log.Println("sec", sec)
			//log.Println("sec e", fmt.Sprintf("%e", sec))
			if pctArray[sec] != 0 {
				result = append(result, map[int64]int64{i: pctArray[sec]})
			}
		}
		i++
	}
	return result
}

func GetPctValue(key int64) string {
	arr := fillPct()
	log.Debug("fillPct", arr)
	var i int64
	for _, pct := range arr {
		if i == key {
			for _, sec := range pct {
				return sec
			}
		}
		i++
	}
	return ""
}

func GetPctArray() [391]map[string]string {
	return fillPct()
}

func FindUserPct(maxUserPctY int) int64 {
	PctArray := GetPctArray()
	var i int64
	for _, pct := range PctArray {
		for PctY, _ := range pct {
			if StrToInt(PctY) >= maxUserPctY {
				if i > 0 {
					return (i - 1)
				} else {
					return 0
				}
			}
		}
		i++
	}
	return 0
}

func DelUserPct(pctArr []map[int64]int64, userMaxKey int64) []map[int64]int64 {
	var new []map[int64]int64
	for i := 0; i < len(pctArr); i++ {
		for key, votes := range pctArr[i] {
			if key > userMaxKey {
				break
			} else {
				new = append(new, map[int64]int64{key: votes})
			}
		}
	}
	return new
}
