// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package lib

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/EGaaS/go-egaas-mvp/packages/test"
)

func TestECDSA(t *testing.T) {
	for i := 0; i < 50; i++ {
		size := rand.Intn(2049)
		if size == 0 {
			continue
		}
		forSign, _ := test.RandBytes(size)
		priv, pub, err := GenBytesKeys()
		if err != nil {
			t.Errorf(err.Error())
		}
		sign, err := SignECDSA(hex.EncodeToString(priv), string(forSign))
		if err != nil {
			t.Errorf(err.Error())
		}
		ret, err := CheckECDSA(pub, string(forSign), sign)
		if err != nil {
			t.Errorf(err.Error())
		}
		if !ret {
			t.Errorf(`ECDSA priv=%x forSign=%x sign=%x`, priv, forSign, sign)
		}
	}
}

func TestJSSign(t *testing.T) {
	data := []test.WantString{
		{`3046022100bf36e83819787c401d454d4471c7fc94513ba3c0fb75b3601d01f828dff83911022100f895df4095e0d4cc8a24434ed5b0647b55b25d6376d92ba3a959c762505fc0f1`,
			`bf36e83819787c401d454d4471c7fc94513ba3c0fb75b3601d01f828dff83911f895df4095e0d4cc8a24434ed5b0647b55b25d6376d92ba3a959c762505fc0f1`},
		{`304502204629ed48c4696793750315ddb460218e3ac2905cb51f00f83730eec0071b0fed0221008cf75ed96e7cbcd6c98842f84c643eb799a89ee26b48846f7eafa55d3ef5860e`,
			`4629ed48c4696793750315ddb460218e3ac2905cb51f00f83730eec0071b0fed8cf75ed96e7cbcd6c98842f84c643eb799a89ee26b48846f7eafa55d3ef5860e`},
		{`304402207441d93243be1fb552baf3b0e832b5e6faadec2b0ea83666ee096325c82bd5ac02203c5b355bc8f7345a6ed528586e4c56309fdb90e839920e0a19d9c2ce56ff5c58`,
			`7441d93243be1fb552baf3b0e832b5e6faadec2b0ea83666ee096325c82bd5ac3c5b355bc8f7345a6ed528586e4c56309fdb90e839920e0a19d9c2ce56ff5c58`},
		{`304402201b68332adf25d17b68c185f82b835aada2216866cab5161228faf581301a6711022038961a5525d513b7f25900222826b86d3d2d0c52dcd0f5c6dca203dbdeefce8c`,
			`1b68332adf25d17b68c185f82b835aada2216866cab5161228faf581301a671138961a5525d513b7f25900222826b86d3d2d0c52dcd0f5c6dca203dbdeefce8c`},
		{`3046022100c26ad7a5a64db475953fcbe54ccf02e943caf86fbc796f0ae5423fe0c4f145d00221009cf9a3cca031f1fc8b0a0c8ba529d1d135459911e3045b3beeb99f6b31a6a3d4`,
			`c26ad7a5a64db475953fcbe54ccf02e943caf86fbc796f0ae5423fe0c4f145d09cf9a3cca031f1fc8b0a0c8ba529d1d135459911e3045b3beeb99f6b31a6a3d4`},
		{`3046023100c26ad7a5a64db475953fcbe54ccf02e943caf86fbc796f0ae5423fe0c4f145d00221009cf9a3cca031f1fc8b0a0c8ba529d1d135459911e3045b3beeb99f6b31a6a3d4`,
			`wrong right parsing`},
	}
	for i, item := range data {
		got, err := JSSignToBytes(item.Input)
		if err != nil {
			if err.Error() != item.Want {
				t.Errorf(`error %s != %v`, item.Want, err)
			}
			continue
		}
		if hex.EncodeToString(got) != item.Want {
			t.Errorf(`%d got=%s want=%s`, i, got, item.Want)
		}
	}
}
