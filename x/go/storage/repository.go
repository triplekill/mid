package storage

// Repository represents APIs for operating database and/or cache
type Repository interface {
	// Name returns database name
	Name() string
	// Insert inserts new records
	Insert(tables ...Table) error
	// Update updates specific fields of record
	Update(table Table, fields ...string) error
	// Find gets records all fields by keys and stores loaded data to container
	Find(meta TableMeta, keys KeyList, container TableListContainer, opts ...GetOption) error
	// FindFields gets records specific fields by keys and stores loaded data to container
	FindFields(meta TableMeta, keys KeyList, container TableListContainer, fields ...string) error
	// Get gets one record all fields
	Get(table Table, opts ...GetOption) (bool, error)
	// GetFields gets one record specific fields
	GetFields(table Table, fields ...string) (bool, error)
	// Remove removes one record
	Remove(table ReadonlyTable) error
	// RemoveRecords removes records by keys
	RemoveRecords(meta TableMeta, keys ...interface{}) error
	// Clear removes all records of table
	Clear(table string) error
	// FindView loads view by keys and stores loaded data to container
	FindView(view View, keys KeyList, container TableListContainer, opts ...GetOption) error
	// IndexRank gets rank of table key in index, returns InvalidRank if key not found
	IndexRank(index Index, key interface{}) (int64, error)
	// IndexScore gets score of table key in index, returns InvalidScore if key not found
	IndexScore(index Index, key interface{}) (int64, error)
	// IndexRange ranges index by rank
	IndexRange(index Index, start, stop int64, opts ...RangeOption) (RangeResult, error)
	// IndexRangeByScore ranges index by score
	IndexRangeByScore(index Index, min, max int64, opts ...RangeOption) (RangeResult, error)
	// IndexRangeByLex ranges index by lexicographical
	IndexRangeByLex(index Index, min, max string, opts ...RangeOption) (RangeLexResult, error)
}
