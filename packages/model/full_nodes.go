package model

type FullNodes struct {
	ID                    int32  `gorm:"primary_key;not_null"`
	Host                  string `gorm:"not null;size:100"`
	WalletID              int64  `gorm:"not null"`
	StateID               int64  `gorm:"not null"`
	FinalDelegateWalletID int64  `gorm:"not null"`
	FinalDelegateStateID  int64  `gorm:"not null"`
	RbID                  int64  `gorm:"not null"`
}

func (fn *FullNodes) FindNode(stateID int64, walletID int64, finalDelegateStateID int64, finalDelegateWalletID int64) error {
	return DBConn.Where(
		"state_id = ?", stateID).Or(
		"wallet_id = ?", walletID).Or(
		"final_delegate_state_id = ?", finalDelegateStateID).Or(
		"final_delegate_wallet_id = ?", finalDelegateWalletID).Find(&fn).Error
}

func (fn *FullNodes) FindNodeByID(nodeID int64) error {
	return DBConn.Where("id = ?", nodeID).First(&fn).Error
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
