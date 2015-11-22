package nanovgo

import (
	"testing"
)

func TestStateInit(t *testing.T) {
	c := Context{}
	c.Save()
	c.Reset()

	topState := c.getState()
	if topState.alpha != 1.0 {
		t.Errorf("initial alpha should be 1.0, but %f", topState.alpha)
	}
}

func TestStateSaveRestore(t *testing.T) {
	c := Context{}
	c.Save()
	c.Reset()

	topState := c.getState()
	topState.alpha = 0.5

	c.Save()

	nextState := c.getState()
	if nextState.alpha != 0.5 {
		t.Errorf("initial alpha should be same with parent's one, but %f", nextState.alpha)
	}
	nextState.alpha = 0.75

	c.Restore()

	topStateAgain := c.getState()
	if topStateAgain.alpha != 0.5 {
		t.Errorf("Restore() should set saved alpha, but %f", topStateAgain.alpha)
	}
}

func equal(a1, a2 TransformMatrix) bool {
	for i := 0; i < 6; i++ {
		if a1[i] != a2[i] {
			return false
		}
	}
	return true
}

func TestStateSaveRestore2(t *testing.T) {
	c := Context{}
	c.Save()
	c.Reset()

	topState := c.getState()
	topState.xform = TranslateMatrix(10, 5)

	c.Save()

	nextState := c.getState()
	if !equal(nextState.xform, TranslateMatrix(10, 5)) {
		t.Errorf("initial xform should be same with parent's one, but %v", nextState.xform)
	}
	nextState.xform = ScaleMatrix(20, 30)

	c.Restore()

	topStateAgain := c.getState()
	if !equal(topStateAgain.xform, TranslateMatrix(10, 5)) {
		t.Errorf("Restore() should set saved xform, but %v", topStateAgain.xform)
	}
}
