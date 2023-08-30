package server

import (
	"testing"
)

func TestWithAddr(t *testing.T) {
	s := &Server{}

	WithAddr("cool-addr")(s)

	if s.addr != "cool-addr" {
		t.Errorf("WithAddr('cool-addr') returned %s", s.addr)
	}
}

func TestWithTempLocation(t *testing.T) {
	s := &Server{}

	WithTempLocation("cool-location")(s)

	if s.tmpLocation != "cool-location" {
		t.Errorf("WithTempLocation('cool-location') returned %s", s.tmpLocation)
	}
}

func TestWithChunkSize(t *testing.T) {
	s := &Server{}

	WithChunkSize(10)(s)

	if s.chunkSize != 10 {
		t.Errorf("WithChunkSize(10) returned %d", s.chunkSize)
	}
}
