package commit

import "testing"

func TestBuildMessage_ConventionalHeader(t *testing.T) {
	c := &ConventionalCommit{
		Subject: "feat(core): add feature",
		Body:    "details",
		Footer:  "Refs: #1",
	}
	got := BuildMessage(c, false, 72)
	want := "feat(core): add feature\n\ndetails\n\nRefs: #1"
	if got != want {
		t.Fatalf("unexpected message:\n%s", got)
	}
}

func TestBuildMessage_Format(t *testing.T) {
	c := &ConventionalCommit{
		Type:    "fix",
		Scope:   "api",
		Subject: "handle nil response",
	}
	got := BuildMessage(c, false, 72)
	want := "fix(api): handle nil response"
	if got != want {
		t.Fatalf("unexpected message:\n%s", got)
	}
}

func TestBuildMessage_EmptyType(t *testing.T) {
	c := &ConventionalCommit{
		Type:    "",
		Subject: "add feature",
		Body:    "details",
		Footer:  "Refs: #1",
	}
	got := BuildMessage(c, false, 72)
	want := "add feature\n\ndetails\n\nRefs: #1"
	if got != want {
		t.Fatalf("unexpected message:\n%s", got)
	}
}
