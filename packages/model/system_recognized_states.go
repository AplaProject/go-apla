package model

type SystemRecognizedStates struct {
	Name             string `gorm:"not null;size:255"`
	StateID          int64  `gorm:"not null;primary_key"`
	Host             string `gorm:"not null;size:255"`
	NodePublickKey   []byte `gorm:"not null"`
	DelegateWalletID int64  `gorm:"not null"`
	DelegateStateID  int64  `gorm:"not null"`
	RbID             int64  `gorm:"not null"`
}

func (srs *SystemRecognizedStates) GetState(stateID int64) error {
	return DBConn.Where("state_id = ?", stateID).First(&srs).Error
}

func (srs *SystemRecognizedStates) IsDelegated(stateID int64) (bool, error) {
	if err := srs.GetState(stateID); err != nil {
		return false, err
	}
	return srs.DelegateStateID > 0 || srs.DelegateWalletID > 0, nil
}
