package tcpserver

import (
	"flag"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/op/go-logging"
	"net"
	"runtime"
	"sync"
)

var (
	log     = logging.MustGetLogger("tcpserver")
	counter int64
	mutex   = &sync.Mutex{}
)

func init() {
	flag.Parse()
}

type TcpServer struct {
	*utils.DCDB
	Conn      net.Conn
}

func (t *TcpServer) deferClose() {
	t.Conn.Close()
	mutex.Lock()
	counter--
	fmt.Println("--", counter)
	mutex.Unlock()
}

func (t *TcpServer) HandleTcpRequest() {

	fmt.Println("NumCPU:", runtime.NumCPU(),
		" NumGoRoutine:", runtime.NumGoroutine(),
		" t.counter:", counter)

	var err error

	log.Debug("HandleTcpRequest from %v", t.Conn.RemoteAddr())
	defer t.deferClose()

	mutex.Lock()
	if counter > 20 {
		t.Conn.Close()
		mutex.Unlock()
		return
	} else {
		counter++
		fmt.Println("++", counter)
	}
	mutex.Unlock()

	// тип данных
	buf := make([]byte, 2)
	_, err = t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	dataType := utils.BinToDec(buf)
	log.Debug("dataType %v", dataType)
	switch dataType {
	case 1:
		t.Type1()
	case 2:
		t.Type2()
	case 4:
		t.Type4()
	case 7:
		t.Type7()
	case 10:
		t.Type10()
	}
	log.Debug("END")
}
