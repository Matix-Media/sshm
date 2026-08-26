package core

import (
	"reflect"
	"testing"
)

var sample = []Connection{
	{ID: "0cce1baf", Name: "colima", Host: "127.0.0.1", User: "max", Port: "50738", IdentityFile: "/keys/id"},
	{ID: "e45a27ab", Name: "VPS 1", Host: "5.75.132.155", User: "matix", Port: "22"},
	{ID: "aaa", Name: "custom", CustomCommand: "mosh me@box"},
}

func TestResolveByID(t *testing.T) {
	c, err := Resolve(sample, "e45a27ab")
	if err != nil || c.Name != "VPS 1" {
		t.Fatalf("got %+v, err %v", c, err)
	}
}

func TestResolveByNameCaseInsensitive(t *testing.T) {
	c, err := Resolve(sample, "COLIMA")
	if err != nil || c.ID != "0cce1baf" {
		t.Fatalf("got %+v, err %v", c, err)
	}
}

func TestResolveIDBeatsName(t *testing.T) {
	conns := []Connection{
		{ID: "shared", Name: "other"},
		{ID: "zzz", Name: "shared"},
	}
	c, err := Resolve(conns, "shared")
	if err != nil || c.Name != "other" {
		t.Fatalf("id should win: got %+v, err %v", c, err)
	}
}

func TestResolveNoMatch(t *testing.T) {
	if _, err := Resolve(sample, "nope"); err == nil {
		t.Fatal("expected error for unknown ref")
	}
}

func TestBuildCommand(t *testing.T) {
	cases := []struct {
		name  string
		conn  Connection
		extra []string
		want  []string
	}{
		{
			name: "default port omits -p",
			conn: Connection{Host: "h", User: "u", Port: "22"},
			want: []string{"ssh", "u@h"},
		},
		{
			name: "non-default port and identity",
			conn: sample[0],
			want: []string{"ssh", "-p", "50738", "-i", "/keys/id", "max@127.0.0.1"},
		},
		{
			name:  "extra args appended",
			conn:  Connection{Host: "h", User: "u", Port: "22"},
			extra: []string{"--", "tic", "-x", "-"},
			want:  []string{"ssh", "u@h", "--", "tic", "-x", "-"},
		},
		{
			name:  "custom command with extra",
			conn:  sample[2],
			extra: []string{"ls"},
			want:  []string{"mosh", "me@box", "ls"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildCommand(tc.conn, tc.extra); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("BuildCommand = %#v, want %#v", got, tc.want)
			}
		})
	}
}
