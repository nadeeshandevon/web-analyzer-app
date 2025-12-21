package metrics

const (
	CategoryValidation = "validation"
	CategoryAuth       = "auth"
	CategoryNotFound   = "not_found"
	CategoryDownstream = "downstream"
	CategoryUnknown    = "unknown"
)

const (
	LabelMethod   = "method"
	LabelEndpoint = "endpoint"
	LabelStatus   = "status"
	LabelCategory = "category"
	LabelReason   = "reason"
)

var APILabels = []string{
	LabelMethod,
	LabelEndpoint,
	LabelStatus,
}
