package model

type BadBlocks struct {
	ID             int64
	ProducerNodeId int64
	BlockId        int64
	ConsumerNodeId int64
	Deleted        bool
}

// TableName returns name of table
func (r BadBlocks) TableName() string {
	return "1_bad_blocks"
}

// BanRequests represents count of unique ban requests for node
type BanRequests struct {
	ProducerNodeId int64
	Count          int64
}

// GetNeedToBanNodes is returns list of ban requests for each node
func (r *BadBlocks) GetNeedToBanNodes() ([]BanRequests, error) {
	var res []BanRequests
	err := DBConn.
		Table(r.TableName()).
		Select("producer_node_id, COUNT(DISTINCT consumer_node_id)").
		Group("producer_node_id").
		Where("deleted = ?", false).
		Scan(&res).
		Error

	return res, err
}

func (r *BadBlocks) GetNodeBlocks(nodeId int64) ([]BadBlocks, error) {
	var res []BadBlocks
	err := DBConn.
		Table(r.TableName()).
		Model(&BadBlocks{}).
		Where("producer_node_id = ? AND deleted = ?", nodeId, false).
		Scan(&res).
		Error

	return res, err
}
