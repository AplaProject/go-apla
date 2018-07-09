package tcpclient

import (
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/network"
)

func SendStopNetwork(addr string, req *network.StopNetworkRequest) error {
	conn, err := newConnection(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	rt := &network.RequestType{
		Type: network.RequestTypeStopNetwork,
	}

	if err = rt.Write(conn); err != nil {
		return err
	}

	if err = req.Write(conn); err != nil {
		return err
	}

	res := &network.StopNetworkResponse{}
	if err = res.Read(conn); err != nil {
		return err
	}

	if len(res.Hash) != consts.HashSize {
		return network.ErrNotAccepted
	}

	return nil
}
