package daemons

import (
	"testing"
	//"github.com/DayLightProject/go-daylight/packages/utils"
	//"github.com/astaxie/beego/config"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DayLightProject/go-daylight/packages/utils"
)

func TestDisseminator(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(dir)

	/*configIni_, err := config.NewConfig("ini", "/home/z/IdeaProjects/src/github.com/DayLightProject/go-daylight/config.ini")
	if err != nil {
		t.Error("%v", utils.ErrInfo(err))
	}
	configIni, err := configIni_.GetSection("default")
	if err != nil {
		t.Error("%v", utils.ErrInfo(err))
	}*/
	host := "pool.dcoin.club:8088"
	userId := int64(2)
	myUserId := int64(4)
	nodePublicKey := utils.HexToBin([]byte("30820122300d06092a864886f70d01010105000382010f003082010a0282010100b85bf8eb06d70e9f28961725ec04957d9c42db127bb788623b5063b7152bdf0df9f1af08a3cdb89f354fe86c43f9f071614b75ccee04ee9e61c749f1800870ad0ada6fc9dbcb928b2049376a06ac6754f6d2832a865e2e4bcfbd1451deb6d2c1ce6a0000701bdce2ec5c20da33ea427a58e9d9bd2807e0c712676593231410b6b0a35b392693cd62e33378987db36b4549ef5f65b172afd7cca9daed6d23e5239d966de9f31a83df4b59cb67ac5c1a816ee633cfcd3a87897b6a9053f3bd4131a403e4a20f301eea5efd31803cdf468663a605cdea96cf6b1cb698b7bb38ab5feb93b68972589d22b910520aab3b20575f2d0bc28b4960b8f912f5b15cede0af0203010001"))
	tx := utils.HexToBin([]byte("0c55b20550013401320131013204302e3031013001300130013001300130013001300130013082010096a9486eb64fd6ca5992e96e879e60881941d7c7bc62a0b86d60d5662b3023e1b8206105cd605791019d01ebc4c0b843284ab58efb772a159066f5a635b94c4344f09f640f244d8f68264cc1c9f83b2471547504041f8c16d8e2af77b07c5fa3799c40f267b1c7fc03326195737b3c605481e4ff37713931c28bc258a83963abf3222c287346b6ba872163b63a676ba9538f6d73fac5ee90500068541c07abddc77dff14eb3a18e47b4157228fe435c79cfa2a6189cef97fcc5f9fd58d1efa4c12f3de3db2f1993d4cd029cd3471f8a82341f75df61af247b70661ae8848afc18d28ab7654f67591f271d826f3925a4b7798653651f1e8c62854ef0bf97127f182010382010056ab8e00429fd794d2b0b64dce4cb0e36d34e59090379249990bcdc824251f907fca3e912a2e1e759dd5622b42aa86149740d72faf9743262875470b57c8c6a8333358a043ec9d8033f799578547bb04eb0f09c8e51898990f4f6760af5213ff61a95506f8b294b0ded892cfa9fcdb801887d1bb405f99ce3656818ce23de0a675cf190a5b616bea7c301f1a76e3dbbfb2c576580daf49e8f83e2286f0a62e869e49295f8bb07da0c25ac0b0a1fd3e7a82887c34f7fdda2112f334e19d7c68a130447536543752ef5a51c16ae456eaf0aadc28dbabcc27e13b7c50942704258e018f6cb898fc4a3c8a7018ee9e529b5f789973b941de7336ce7bf08b79cfeae7"))
	hash := utils.HexToBin(utils.Md5(tx))
	toBeSent := utils.DecToBin(myUserId, 5)
	toBeSent = append(toBeSent, utils.DecToBin(1, 1)...)
	toBeSent = append(toBeSent, utils.DecToBin(utils.StrToInt64("0"), 1)...)
	toBeSent = append(toBeSent, []byte(hash)...)
	fmt.Printf("hash %x %v", hash, hash)
	dataType := int64(1)
	db := DbConnect(chBreaker, chAnswer, GoroutineName)
	DisseminatorType1(host, userId, string(nodePublicKey), db, toBeSent, dataType)

}
