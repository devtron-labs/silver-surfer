package kubedd

import (
	"github.com/devtron-labs/silver-surfer/pkg"
	"reflect"
	"testing"
)

func TestValidateCluster(t *testing.T) {
	cluster := pkg.NewCluster("", "")
	config := pkg.NewDefaultConfig()
	config.SelectKinds = []string{"ComponentStatus"}
	config.TargetKubernetesVersion = "1.21"
	//config.SelectNamespaces = []string{"prod"}
	type args struct {
		cluster *pkg.Cluster
		conf    *pkg.Config
	}
	tests := []struct {
		name    string
		args    args
		want    []pkg.ValidationResult
		wantErr bool
	}{
		{
			name: "check cluster",
			args: args{
				cluster: cluster,
				conf:    config,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateCluster(tt.args.cluster, tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateCluster() got = %v, want %v", got, tt.want)
			}
		})
	}
}
