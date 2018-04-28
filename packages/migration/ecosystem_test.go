package migration

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestGetEcosystemScript(t *testing.T) {
	str := fmt.Sprintf(GetFirstEcosystemScript(), -1744264011260937456)
	ioutil.WriteFile("/home/losaped/ecosystem_test.sql", []byte(str), 0777)
}
