package webhook

import "testing"

func TestImagesBaseEqual(t *testing.T) {
	if !imagesBaseEqual("registry.xxx.com/namespace/aaa:master-1", "registry.xxx.com/namespace/aaa:master-22") {
		t.Fail()
	}

	if imagesBaseEqual("registry.xxx.com/namespace/aaa:master-1", "registry.xxx.com/namespace/bbbb:master-2") {
		t.Fail()
	}

	if imagesBaseEqual("registry.xxx.com/namespace/aaa:master-1", "registry.xxx.com/namespace/aaa:zzzz-2") {
		t.Fail()
	}
}
