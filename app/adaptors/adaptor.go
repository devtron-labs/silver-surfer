package adaptors

import (
	"github.com/devtron-labs/silver-surfer/app/grpc"
	"github.com/devtron-labs/silver-surfer/pkg"
)

func ConvertSummaryValidationResultToGrpcObj(req []pkg.SummaryValidationResult) []*grpc.SummaryValidationResult {
	resp := make([]*grpc.SummaryValidationResult, len(req))
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
	resp := make([]*grpc.SummarySchemaError, len(req))
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
