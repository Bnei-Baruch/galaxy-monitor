package api

import (
	"fmt"

	"github.com/pkg/errors"
)

func handleSpecPost(r Spec) *HttpError {
	if r.SampleInterval < SECOND_MS {
		return NewBadRequestError(errors.New(fmt.Sprintf("sample_interval smaller then second, got %d", r.SampleInterval)))
	}
	if r.StoreInterval < MINUTE_MS {
		return NewBadRequestError(errors.New(fmt.Sprintf("store_interval smaller then minute, got %d", r.StoreInterval)))
	}

	SPEC.SampleInterval = r.SampleInterval
	SPEC.StoreInterval = r.StoreInterval
	if len(r.MetricsWhitelist) > 0 {
		SPEC.MetricsWhitelist = r.MetricsWhitelist
	}

	return nil
}

func handleSpecGet() (*Spec, *HttpError) {
	return &SPEC, nil
}
