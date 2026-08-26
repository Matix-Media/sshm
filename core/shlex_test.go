package core

import (
	"reflect"
	"testing"
)

func TestShlexSplit(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"ssh user@host", []string{"ssh", "user@host"}},
		{"a   b\tc", []string{"a", "b", "c"}},
		{`ssh -o "ProxyCommand cloudflared access ssh"`, []string{"ssh", "-o", "ProxyCommand cloudflared access ssh"}},
		{`echo 'hello world'`, []string{"echo", "hello world"}},
		{`mix'ed'quotes`, []string{"mixedquotes"}},
		{`unterminated "quote here`, []string{"unterminated", "quote here"}},
	}
	for _, tc := range cases {
		if got := shlexSplit(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("shlexSplit(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}
