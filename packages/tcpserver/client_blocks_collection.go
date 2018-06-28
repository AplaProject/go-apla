package tcpserver

import (
	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
)

// GetBlocksBodies send GetBodiesRequest returns channel of binary blocks data
func GetBlocksBodies(host string, blockID int64, reverseOrder bool) (chan []byte, error) {
	conn, err := newConnection(host)
	if err != nil {
		return nil, err
	}

	// send the type of data
	rt := &RequestType{Type: RequestTypeBlockCollection}
	if err = rt.Write(conn); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data type block body to connection")
		return nil, err
	}

	req := &GetBodiesRequest{
		BlockID:      blockID,
		ReverseOrder: reverseOrder,
	}

	if err = req.Write(con); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending blocks bodies request")
		return nil, err
	}

	rawBlocksCh := make(chan []byte, tcpserver.BlocksPerRequest)
	go func() {
		defer func() {
			close(rawBlocksCh)
			conn.Close()
		}()

		for {
			// receive the data size as a response that server wants to transfer
			resp := &GetBodyResponse{}
			if err := resp.Read(conn); err != nil {
				log.WithFields(log.Fields{"type": contsts.IOError, "error": err}).Error("on reading Block body response")
				return
			}

			// TODO: remove fucking hardcode
			// TODO: move size checking to GetBodyResponse.Read with limitReader
			// data size must be less than 10mb
			if resp.Data > 10485760 || dataSize == 0 {
				log.Error("null block")
				return
			}

			rawBlocksCh <- binaryBlock
		}
	}()
	return rawBlocksCh, nil

}
