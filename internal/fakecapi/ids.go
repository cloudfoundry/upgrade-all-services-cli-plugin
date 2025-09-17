package fakecapi

import (
	"crypto/sha256"
	"fmt"
)

// fakeName returns a test fake name for the specified resource kind
func (f *FakeCAPI) fakeName(kind string) string {
	f.fakeNameCount[kind]++
	return fmt.Sprintf("fake-%s-%d", kind, f.fakeNameCount[kind])
}

// stableGUID returns a GUID that is a function of the specified string rather than random.
// This allows GUIDs to be included in expected matches in tests.
func stableGUID(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	b := h.Sum(nil)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
