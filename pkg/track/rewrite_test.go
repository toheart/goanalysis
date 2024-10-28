package track

import (
	"testing"
)

/**
@file:
@author: levi.Tang
@time: 2024/10/28 19:52
@description:
**/

func TestRewrite(t *testing.T) {
	r, err := NewRewrite("../../example/main.go")
	if err != nil {
		t.Errorf("new Rewrite err:%s", err)
		return
	}

	t.Run("TestHasImportFunctrace", func(t *testing.T) {
		r.ImportFunctrace()
	})
}
