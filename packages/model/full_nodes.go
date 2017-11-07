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

func (fn *FullNode) FindNode(stateID int64, walletID int64, finalDelegateStateID int64, finalDelegateWalletID int64) (bool, error) {
	return isFound(DBConn.Where(
		"state_id = ?", stateID).Or(
		"wallet_id = ?", walletID).Or(
		"final_delegate_state_id = ?", finalDelegateStateID).Or(
		"final_delegate_wallet_id = ?", finalDelegateWalletID).Find(&fn))
}

func (fn *FullNode) Get(walletID int64) (bool, error) {
	return isFound(DBConn.Where("wallet_id = ?", walletID).First(fn))
}

func (fn *FullNode) FindNodeByID(nodeID int64) (bool, error) {
	return isFound(DBConn.Where("id = ?", nodeID).First(fn))
}

func (fn *FullNode) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(fn).Error
}

// TODO: delete full_nodes table
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
	err := DBConn.Find(&nodes).Error
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
