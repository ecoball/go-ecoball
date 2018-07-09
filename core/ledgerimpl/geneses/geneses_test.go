package geneses_test

import (
	"testing"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/geneses"
)

func TestGenesesBlockInit(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/geneses")
	if err != nil {
		t.Fatal(err)
	}
	block, err := geneses.GenesesBlockInit(l)
	if err != nil {
		t.Fatal(err)
	}
	block.Show()
}
