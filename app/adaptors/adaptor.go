package adaptors

import (
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/common-lib/utils/remoteConnection/bean"
	"github.com/devtron-labs/silver-surfer/app/grpc"
	"github.com/devtron-labs/silver-surfer/pkg"
)

func ConvertSummaryValidationResultToGrpcObj(req []pkg.SummaryValidationResult) []*grpc.SummaryValidationResult {
	resp := make([]*grpc.SummaryValidationResult, 0, len(req))
	for _, item := range req {
		svr := &grpc.SummaryValidationResult{
			FileName:               item.FileName,
			Kind:                   item.Kind,
			APIVersion:             item.APIVersion,
			ResourceName:           item.ResourceName,
			ResourceNamespace:      item.ResourceNamespace,
			Deleted:                item.Deleted,
			Deprecated:             item.Deprecated,
			LatestAPIVersion:       item.LatestAPIVersion,
			IsVersionSupported:     int32(item.IsVersionSupported),
			ErrorsForOriginal:      ConvertSummarySchemaErrorToGrpcObj(item.ErrorsForOriginal),
			ErrorsForLatest:        ConvertSummarySchemaErrorToGrpcObj(item.ErrorsForLatest),
			DeprecationForOriginal: ConvertSummarySchemaErrorToGrpcObj(item.DeprecationForOriginal),
			DeprecationForLatest:   ConvertSummarySchemaErrorToGrpcObj(item.DeprecationForLatest),
		}
		resp = append(resp, svr)
	}
	return resp
}

func ConvertSummarySchemaErrorToGrpcObj(req []*pkg.SummarySchemaError) []*grpc.SummarySchemaError {
	resp := make([]*grpc.SummarySchemaError, 0, len(req))
	for _, item := range req {
		if item != nil {
			sse := &grpc.SummarySchemaError{
				Path:        item.Path,
				SchemaField: item.SchemaField,
				Reason:      item.Reason,
			}
			resp = append(resp, sse)
		}
	}
	return resp
}

func ConvertGrpcObjToClusterConfig(req *grpc.ClusterConfig) *k8s.ClusterConfig {
	if req != nil {
		return &k8s.ClusterConfig{
			ClusterName:            req.ClusterName,
			Host:                   req.ApiServerUrl,
			BearerToken:            req.Token,
			InsecureSkipTLSVerify:  req.InsecureSkipTLSVerify,
			KeyData:                req.KeyData,
			CertData:               req.CertData,
			CAData:                 req.CaData,
			ClusterId:              int(req.ClusterId),
			RemoteConnectionConfig: ConvertGrpcObjToRemoteConnectionConfig(req.RemoteConnectionConfig),
		}
	}
	return &k8s.ClusterConfig{}
}

func ConvertGrpcObjToRemoteConnectionConfig(req *grpc.RemoteConnectionConfig) *bean.RemoteConnectionConfigBean {
	if req != nil {
		return &bean.RemoteConnectionConfigBean{
			ConnectionMethod: bean.RemoteConnectionMethod(req.RemoteConnectionMethod),
			ProxyConfig:      ConvertGrpcObjToProxyConfig(req.ProxyConfig),
			SSHTunnelConfig:  ConvertGrpcObjToSSHTunnelConfig(req.SSHTunnelConfig),
		}
	}
	return &bean.RemoteConnectionConfigBean{}
}

func ConvertGrpcObjToProxyConfig(req *grpc.ProxyConfig) *bean.ProxyConfig {
	if req != nil {
		return &bean.ProxyConfig{ProxyUrl: req.ProxyUrl}
	}
	return &bean.ProxyConfig{}
}

func ConvertGrpcObjToSSHTunnelConfig(req *grpc.SSHTunnelConfig) *bean.SSHTunnelConfig {
	if req != nil {
		return &bean.SSHTunnelConfig{
			SSHServerAddress: req.SSHServerAddress,
			SSHUsername:      req.SSHUsername,
			SSHPassword:      req.SSHPassword,
			SSHAuthKey:       req.SSHAuthKey,
		}
	}
	return &bean.SSHTunnelConfig{}
}
