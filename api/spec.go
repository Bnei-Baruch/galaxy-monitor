package api

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

func handleSpecPost(db *sql.DB, r Spec) (*EmptyResponse, *HttpError) {
	if r.SampleInterval >= SECOND_MS {
		SPEC.SampleInterval = r.SampleInterval
	} else {
		return &EmptyResponse{}, NewBadRequestError(errors.New(fmt.Sprintf("sample_interval smaller then second, got %d", r.SampleInterval)))
	}
	if r.StoreInterval >= MINUTE_MS {
		SPEC.StoreInterval = r.StoreInterval
	} else {
		return &EmptyResponse{}, NewBadRequestError(errors.New(fmt.Sprintf("store_interval smaller then minute, got %d", r.StoreInterval)))
	}
	if len(r.MetricsBlacklist) > 0 {
		SPEC.MetricsBlacklist = r.MetricsBlacklist
	}
	return &EmptyResponse{}, nil
}

func handleSpecGet(db *sql.DB) (*Spec, *HttpError) {
	return &SPEC, nil
}
