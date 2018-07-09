package bijection_test

import (
	"testing"
	"github.com/eager7/go_study/2018/bijection"
)

func TestNew(t *testing.T) {
	m := bijection.New()
	if err := m.Set(1, 3); err != nil {
		t.Fatal(err)
	}
	if err := m.Set(2, 2); err != nil {
		t.Fatal(err)
	}
}
