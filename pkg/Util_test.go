package pkg

import (
	"fmt"
	"testing"
)

func TestRegexMatch(t *testing.T) {
	type args struct {
		s       string
		pattern string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ends with match",
			args: args{
				s:       "spec/jobTemplate/spec/template/metadata/creationTimestamp",
				pattern: "*creationTimestamp",
			},
			want: true,
		},
		{
			name: "starts with match",
			args: args{
				s:       "spec/jobTemplate/spec/template/metadata/creationTimestamp",
				pattern: "spec*",
			},
			want: true,
		},
		{
			name: "has match",
			args: args{
				s:       "spec/jobTemplate/spec/template/metadata/creationTimestamp",
				pattern: "*job*",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrationStatus := fmt.Sprintf("%s%d%s%s%s", "\033[31m", 1, " issue(s):", "\033[100;97m", " fix issues before migration")

			fmt.Printf("|%6s|%6s|\n", "foo", "b")
			fmt.Printf("|%6s|%6s|\n", "foo", "\033[97m bn")
			fmt.Println(migrationStatus)
			if got := RegexMatch(tt.args.s, tt.args.pattern); got != tt.want {
				t.Errorf("RegexMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
