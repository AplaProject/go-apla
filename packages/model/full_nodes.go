package model

import "strconv"

type FullNode struct {
	ID                    int32  `gorm:"primary_key;not_null"`
	Host                  string `gorm:"not null;size:100"`
	WalletID              int64  `gorm:"not null default 0"`
	StateID               int64  `gorm:"not null default 0"`
	FinalDelegateWalletID int64  `gorm:"not null default 0"`
	FinalDelegateStateID  int64  `gorm:"not null default 0"`
	RbID                  int64  `gorm:"not null default 0"`
}

func (fn *FullNode) FindNode(stateID int64, walletID int64, finalDelegateStateID int64, finalDelegateWalletID int64) error {
	return handleError(DBConn.Where(
		"state_id = ?", stateID).Or(
		"wallet_id = ?", walletID).Or(
		"final_delegate_state_id = ?", finalDelegateStateID).Or(
		"final_delegate_wallet_id = ?", finalDelegateWalletID).Find(&fn).Error)
}

func (fn *FullNode) Get(walletID int64) error {
	return handleError(DBConn.Where("wallet_id = ?", walletID).First(fn).Error)
}

func (fn *FullNode) FindNodeByID(nodeID int64) error {
	return handleError(DBConn.Where("id = ?", nodeID).First(&fn).Error)
}

func (fn *FullNode) GetAllFullNodesHasWalletID() ([]FullNode, error) {
	result := make([]FullNode, 0)
	err := DBConn.Where("wallet_id != 0").Find(&result).Error
	return result, err
}

func (fn *FullNode) GetRbIDFullNodesWithWallet() error {
	return handleError(DBConn.Where("wallet_id != 0").First(fn).Error)
}

func (fn *FullNode) DeleteNodesWithWallets() error {
	return DBConn.Exec("DELETE FROM full_nodes WHERE wallet_id != 0").Error
}

func (fn *FullNode) FindNodeById(nodeid int64) error {
	return handleError(DBConn.Where("id = ?", nodeid).First(&fn).Error)
}

func (fn *FullNode) Create() error {
	return DBConn.Create(fn).Error
}

func FullNodeCreateTable() error {
	return DBConn.CreateTable(&FullNode{}).Error
}

func GetFullNodesHosts() ([]string, error) {
	hosts := new([]string)
	rows, err := DBConn.Table("full_nodes").Select("DISTINCT ON (host) host").Rows()
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	for rows.Next() {
		var host string
		if err := rows.Scan(&host); err != nil {
			return nil, nil
		}
		*hosts = append(*hosts, host)
	}
	if err := rows.Err(); err != nil {
		return nil, nil
	}
	return *hosts, nil
}

func (fn *FullNode) GetAll() (*[]FullNode, error) {
	nodes := new([]FullNode)
	err := DBConn.Find(nodes).Error
	return nodes, err
}

func (fn *FullNode) ToMap() map[string]string {
	result := make(map[string]string)
	result["id"] = strconv.FormatInt(int64(fn.ID), 10)
	result["host"] = fn.Host
	result["wallet_id"] = strconv.FormatInt(fn.WalletID, 10)
	result["state_id"] = strconv.FormatInt(fn.StateID, 10)
	result["final_delegate_wallet_id"] = strconv.FormatInt(fn.FinalDelegateWalletID, 10)
	result["final_delegate_state_id"] = strconv.FormatInt(fn.FinalDelegateStateID, 10)
	result["rb_id"] = strconv.FormatInt(fn.RbID, 10)
	return result
}

func (fn *FullNode) GetMaxID() (int32, error) {
	var result int32
	err := DBConn.Raw("SELECT max(id) FROM full_nodes").Row().Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}
