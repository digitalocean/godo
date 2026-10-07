package godo

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// Insights logs filter operators.
const (
	InsightsLogsFilterOperatorEqual              = "FILTER_OPERATOR_EQ"
	InsightsLogsFilterOperatorNotEqual           = "FILTER_OPERATOR_NEQ"
	InsightsLogsFilterOperatorIn                 = "FILTER_OPERATOR_IN"
	InsightsLogsFilterOperatorExists             = "FILTER_OPERATOR_EXISTS"
	InsightsLogsFilterOperatorGreaterThanOrEqual = "FILTER_OPERATOR_GTE"
	InsightsLogsFilterOperatorLessThanOrEqual    = "FILTER_OPERATOR_LTE"
)

// Insights logs field scopes.
const (
	InsightsLogsFieldScopeResource   = "FIELD_SCOPE_RESOURCE"
	InsightsLogsFieldScopeAttributes = "FIELD_SCOPE_ATTRIBUTES"
)

// Insights logs sort directions.
const (
	InsightsLogsSortDirectionAscending  = "SORT_DIRECTION_ASC"
	InsightsLogsSortDirectionDescending = "SORT_DIRECTION_DESC"
)

// LogsSearchRequest describes an Insights logs search.
type LogsSearchRequest struct {
	TimeRange  LogsTimeRange          `json:"time_range"`
	Filter     *LogsFilterExpression  `json:"filter,omitempty"`
	OrderBy    []LogsOrderBy          `json:"order_by,omitempty"`
	Pagination *LogsPaginationRequest `json:"pagination,omitempty"`
}

// LogsTimeRange is an inclusive logs query time window.
type LogsTimeRange struct {
	From LogsTimeInstant `json:"from"`
	To   LogsTimeInstant `json:"to"`
}

// LogsTimeInstant identifies a point in time. Set exactly one of Absolute,
// Relative, or UnixNano.
type LogsTimeInstant struct {
	Absolute string `json:"absolute,omitempty"`
	Relative string `json:"relative,omitempty"`
	UnixNano string `json:"unix_nano,omitempty"`
}

// LogsFilterExpression is a boolean filter tree. Set exactly one expression
// node.
type LogsFilterExpression struct {
	Condition  *LogsFilterCondition  `json:"condition,omitempty"`
	And        *LogsFilterGroup      `json:"and,omitempty"`
	Or         *LogsFilterGroup      `json:"or,omitempty"`
	Not        *LogsFilterExpression `json:"not,omitempty"`
	TextSearch *LogsTextSearch       `json:"text_search,omitempty"`
}

// LogsFilterGroup combines nested logs filter expressions.
type LogsFilterGroup struct {
	Expressions []LogsFilterExpression `json:"expressions"`
}

// LogsTextSearch searches for a substring across the log body, service name,
// and resource URN.
type LogsTextSearch struct {
	Query string `json:"query"`
}

// LogsFilterCondition is a single field comparison.
type LogsFilterCondition struct {
	Field    LogsFieldRef     `json:"field"`
	Operator string           `json:"operator"`
	Value    *LogsFilterValue `json:"value,omitempty"`
}

// LogsFieldRef identifies an intrinsic field or an attribute.
type LogsFieldRef struct {
	Name  string `json:"name"`
	Scope string `json:"scope,omitempty"`
}

// LogsFilterValue is a typed filter literal. Set exactly one value.
// Scalar fields use pointers so zero, false, and an empty string can be sent.
type LogsFilterValue struct {
	StringValue      *string               `json:"string_value,omitempty"`
	NumberValue      *float64              `json:"number_value,omitempty"`
	BoolValue        *bool                 `json:"bool_value,omitempty"`
	StringArrayValue *LogsStringArrayValue `json:"string_array_value,omitempty"`
	NumberArrayValue *LogsNumberArrayValue `json:"number_array_value,omitempty"`
}

// LogsStringArrayValue is a string-list filter value.
type LogsStringArrayValue struct {
	Values []string `json:"values"`
}

// LogsNumberArrayValue is a number-list filter value.
type LogsNumberArrayValue struct {
	Values []float64 `json:"values"`
}

// LogsOrderBy is a logs search sort clause.
type LogsOrderBy struct {
	Field     LogsFieldRef `json:"field"`
	Direction string       `json:"direction,omitempty"`
}

// LogsPaginationRequest configures cursor pagination for a logs search.
type LogsPaginationRequest struct {
	Limit  int    `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

// LogsSearchResponse contains matching log records and pagination state.
type LogsSearchResponse struct {
	Data       []InsightsLogRecord     `json:"data,omitempty"`
	Pagination *LogsPaginationResponse `json:"pagination,omitempty"`
}

// LogsPaginationResponse describes the next page of logs search results.
type LogsPaginationResponse struct {
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// InsightsLogRecord is a single log record returned by an Insights logs search.
type InsightsLogRecord struct {
	Timestamp      time.Time         `json:"timestamp"`
	SeverityNumber int32             `json:"severity_number,omitempty"`
	SeverityText   string            `json:"severity_text,omitempty"`
	Body           string            `json:"body,omitempty"`
	TraceID        string            `json:"trace_id,omitempty"`
	SpanID         string            `json:"span_id,omitempty"`
	ServiceName    string            `json:"service_name,omitempty"`
	Resource       map[string]string `json:"resource,omitempty"`
	Attributes     map[string]string `json:"attributes,omitempty"`
}

// SearchLogs searches log records in a region.
func (s *InsightsServiceOp) SearchLogs(ctx context.Context, region string, search *LogsSearchRequest) (*LogsSearchResponse, *Response, error) {
	path := "/v2/insights/query/" + url.PathEscape(region) + "/logs/search"
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, search)
	if err != nil {
		return nil, nil, err
	}

	result := new(LogsSearchResponse)
	resp, err := s.client.Do(ctx, req, result)
	if err != nil {
		return nil, resp, err
	}
	return result, resp, nil
}
