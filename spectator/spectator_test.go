package spectator_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/spectator"
)

/***********************************************************/
type WhtBlock struct {
	Height     uint64
	CountTxs   uint32
	PrevHash   common.Hash
	MerkleHash common.Hash
	StateHash  common.Hash
	Hash       common.Hash
}

func (this *WhtBlock) Serialize() ([]byte, error) {
	return json.Marshal(*this)
}

func (this *WhtBlock) Deserialize(data []byte) error {
	return json.Unmarshal(data, this)
}

func Testspectator(t *testing.T) {
	go spectator.Bystander()
	sendMessage()
}

func sendMessage() {
	for {
		time.Sleep(2 * time.Second)
		hash := common.Hash{}
		msg := &WhtBlock{
			Height:     2,
			CountTxs:   10,
			PrevHash:   hash.FormHexString("123"),
			MerkleHash: hash.FormHexString("123"),
			StateHash:  hash.FormHexString("123"),
			Hash:       hash.FormHexString("123"),
		}
		spectator.Notify(spectator.InfoBlock, msg)
	}
}
