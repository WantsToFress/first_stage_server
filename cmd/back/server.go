package main

import (
	"crypto/rsa"
	"strings"

	"github.com/centrifugal/centrifuge-go"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	event "github.com/wantsToFress/first_stage_server/pkg"
)

func paginatedQuery(query *orm.Query, pagination *event.Pagination, allowedFields ...string) (*orm.Query, error) {
	if pagination == nil {
		return query, nil
	}

	query.Offset(int(pagination.Offset))
	query.Limit(int(pagination.Limit))

	if pagination.SortBy != "" {
		flag := false
		for i := range allowedFields {
			if pagination.SortBy == allowedFields[i] {
				flag = true
				break
			}
		}
		if !flag {
			return nil, status.Error(codes.InvalidArgument, "invalid sort_by field (permitted value ["+
				strings.Join(allowedFields, ",")+"])")
		}

		if pagination.Order != event.SortOrder_NO_ORDER {
			if pagination.Order != event.SortOrder_DESC && pagination.Order != event.SortOrder_ASC {
				return nil, status.Error(codes.InvalidArgument, "invalid order value (permitted value [desc,asc])")
			}
			query.Order(pagination.SortBy + " " + pagination.Order.String())
		} else {
			query.Order(pagination.SortBy)
		}
	} else {
		if pagination.Order != event.SortOrder_NO_ORDER {
			return nil, status.Error(codes.InvalidArgument, "invalid order value (sort_by field not set)")
		}
	}

	return query, nil
}

func paginationWithHits(p *event.Pagination, totalHits int) *event.Pagination {
	if p != nil {
		return &event.Pagination{
			TotalHits: uint64(totalHits),
			Limit:     p.Limit,
			Offset:    p.Offset,
			SortBy:    p.SortBy,
			Order:     p.Order,
		}
	} else {
		return &event.Pagination{
			TotalHits: uint64(totalHits),
		}
	}
}

func stringToStringWrapper(val string) *wrappers.StringValue {
	return &wrappers.StringValue{Value: val}
}

func stringWrapperToPtr(val *wrappers.StringValue) *string {
	if val == nil || val.GetValue() == "" {
		return nil
	}
	return &val.Value
}

func ptrToStringWrapper(val *string) *wrappers.StringValue {
	if val == nil {
		return nil
	}
	return &wrappers.StringValue{Value: *val}
}

func boolWrapperToPtr(val *wrappers.BoolValue) *bool {
	if val == nil {
		return nil
	}
	return &val.Value
}

func ptrToBoolWrapper(val *bool) *wrappers.BoolValue {
	if val == nil {
		return nil
	}
	return &wrappers.BoolValue{Value: *val}
}

func ptrToBool(val *bool) bool {
	if val == nil {
		return false
	}
	return *val
}

func float64WrapperToPtr(val *wrappers.DoubleValue) *float64 {
	if val == nil {
		return nil
	}
	return &val.Value
}

func ptrToDoubleWrapper(val *float64) *wrappers.DoubleValue {
	if val == nil {
		return nil
	}
	return &wrappers.DoubleValue{Value: *val}
}

func int64WrapperToPtr(val *wrappers.Int64Value) *int64 {
	if val == nil {
		return nil
	}
	return &val.Value
}

func ptrToInt64Wrapper(val *int64) *wrappers.Int64Value {
	if val == nil {
		return nil
	}
	return &wrappers.Int64Value{Value: *val}
}

func int64ToInt64Wrapper(val int64) *wrappers.Int64Value {
	return &wrappers.Int64Value{Value: val}
}

func int32WrapperToPtr(val *wrappers.Int32Value) *int32 {
	if val == nil {
		return nil
	}
	return &val.Value
}

func ptrToInt32Wrapper(val *int32) *wrappers.Int32Value {
	if val == nil {
		return nil
	}
	return &wrappers.Int32Value{Value: *val}
}

func int32ToInt32Wrapper(val int32) *wrappers.Int32Value {
	return &wrappers.Int32Value{Value: val}
}

func boolToBoolWrapper(v bool) *wrappers.BoolValue {
	return &wrappers.BoolValue{Value: v}
}

type boolSelector struct {
	Value *bool
}

func boolToBoolSelector(val bool) *boolSelector {
	return &boolSelector{Value: &val}
}

func ptrToBoolSelector(val *bool) *boolSelector {
	return &boolSelector{Value: val}
}

func boolSelectorToBool(s *boolSelector) bool {
	if s == nil {
		return false
	}
	if s.Value == nil {
		return false
	}
	return *s.Value
}

func boolSelectorToPtr(s *boolSelector) *bool {
	if s == nil {
		return nil
	}
	return s.Value
}

func boolSelectorToBoolWrapper(s *boolSelector) *wrappers.BoolValue {
	if s == nil {
		return nil
	}
	if s.Value == nil {
		return nil
	}
	return &wrappers.BoolValue{Value: *s.Value}
}

func boolWrapperToBoolSelector(s *wrappers.BoolValue) *boolSelector {
	if s == nil {
		return nil
	}
	return &boolSelector{Value: &s.Value}
}

func difference(a, b []string) []string {
	diff := []string{}
	m := make(map[string]struct{})

	for i := range b {
		m[b[i]] = struct{}{}
	}

	for i := range a {
		if _, ok := m[a[i]]; !ok {
			diff = append(diff, a[i])
		}
	}
	return diff
}

func set(a []string) []string {
	m := make(map[string]struct{})
	for i := range a {
		m[a[i]] = struct{}{}
	}
	res := make([]string, 0, len(m))
	for i := range m {
		res = append(res, i)
	}
	return res
}

type EventService struct {
	db         *pg.DB
	cent       *centrifuge.Client
	privateKey rsa.PrivateKey
	publicKey  rsa.PublicKey
}
