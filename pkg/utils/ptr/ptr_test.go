package ptr

import (
	"reflect"
	"testing"
)

func equal(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, actual %#v", expected, actual)
	}
}

func TestOf(t *testing.T) {
	equal(t, int(10), *Of(10))
}

func TestTo(t *testing.T) {
	equal(t, int(10), To(Of(10)))
	equal(t, int(0), To((*int)(nil)))
}

func TestToDef(t *testing.T) {
	equal(t, int(10), ToDef(Of(10), 0))
	equal(t, int(5), ToDef(nil, 5))
}
