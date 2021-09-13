package snowflake

import "time"

// ID a snowflake ID
type ID struct {
	Timestamp  time.Time `json:"timestamp,omitempty"`
	SinceEpoch uint64    `json:"since_epoch,omitempty"`
	NodeID     uint16    `json:"node_id,omitempty"`
	Sequence   uint16    `json:"sequence,omitempty"`
}
