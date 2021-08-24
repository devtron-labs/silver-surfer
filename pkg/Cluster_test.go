package pkg

import (
	"testing"
)

func TestCluster_ServerVersion(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name: "cluster version",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCluster("", "")
			config := NewDefaultConfig()
			config.SelectKinds = []string{"deployment"}
			config.TargetKubernetesVersion = "1.16"
			config.SelectNamespaces = []string{"prod"}
			got, err := c.ServerVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("ServerVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ServerVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}