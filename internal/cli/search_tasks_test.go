package cli

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

func TestSearchTasksBuildsQuery(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	out, err := runWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"gid":"t1","name":"Release","completed":false}],"next_page":null}`))
	}, "search-tasks", "--workspace-gid", "ws1", "--text", "release", "--assignee", "me")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/workspaces/ws1/tasks/search" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery.Get("text") != "release" {
		t.Errorf("text = %q", gotQuery.Get("text"))
	}
	if gotQuery.Get("assignee.any") != "me" {
		t.Errorf("assignee.any = %q", gotQuery.Get("assignee.any"))
	}
	if _, ok := gotQuery["completed"]; ok {
		t.Errorf("completed should be omitted when unset, got %q", gotQuery.Get("completed"))
	}
	var tasks []json.RawMessage
	decodeData(t, out, &tasks)
	if len(tasks) != 1 {
		t.Errorf("got %d tasks", len(tasks))
	}
}

func TestSearchTasksCompletedTriState(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string // "" means key absent
	}{
		{"unset", []string{"search-tasks", "--workspace-gid", "ws1"}, ""},
		{"true", []string{"search-tasks", "--workspace-gid", "ws1", "--completed=true"}, "true"},
		{"false", []string{"search-tasks", "--workspace-gid", "ws1", "--completed=false"}, "false"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotQuery url.Values
			_, err := runWithServer(t, func(w http.ResponseWriter, r *http.Request) {
				gotQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"data":[],"next_page":null}`))
			}, tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, present := gotQuery["completed"]
			if tc.want == "" {
				if present {
					t.Errorf("completed present = %v, want absent", got)
				}
				return
			}
			if !present || got[0] != tc.want {
				t.Errorf("completed = %v, want %q", got, tc.want)
			}
		})
	}
}

func TestSearchTasksPremiumRequiredIsRuntimeError(t *testing.T) {
	_, err := runWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(402)
		w.Write([]byte(`{"errors":[]}`))
	}, "search-tasks", "--workspace-gid", "ws1", "--text", "x")
	if err == nil {
		t.Fatal("expected error")
	}
	if exitCodeFor(err) != exitRuntime {
		t.Errorf("exit code = %d, want %d", exitCodeFor(err), exitRuntime)
	}
	if err.Error() != "Asana API access requires a premium workspace or feature for this request." {
		t.Errorf("message = %q", err.Error())
	}
}
