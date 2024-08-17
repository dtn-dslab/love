package test

import (
	"testing"

	qdisc "github.com/dtn-dslab/love/pkg/qdisc"
	"github.com/vishvananda/netlink"
)

func TestParseQdiscsEmpty(t *testing.T) {
	intfProps := &qdisc.DeviceProperties{}

	qdiscs, err := intfProps.ParseQdiscs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(qdiscs) != 0 {
		t.Fatalf("unexpected qdiscs: %v", qdiscs)
	}
}

func TestParseQdiscsRate(t *testing.T) {
	intfProps := &qdisc.DeviceProperties{
		Rate: "100 mbps",
	}

	qdiscs, err := intfProps.ParseQdiscs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(qdiscs) != 1 {
		t.Fatalf("unexpected qdiscs: %v", qdiscs)
	}

	tbf, ok := qdiscs[0].(*netlink.Tbf)
	if !ok {
		t.Fatalf("unexpected qdisc: %v", qdiscs[0])
	}

	if tbf.Rate != 800000000 {
		t.Fatalf("unexpected rate: %v", tbf.Rate)
	}
}

func TestParseQdiscsLNetem(t *testing.T) {
	intfProps := &qdisc.DeviceProperties{
		Latency: "100ms",
		Gap:     10,
		Loss:    10.0,
	}

	qdiscs, err := intfProps.ParseQdiscs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(qdiscs) != 1 {
		t.Fatalf("unexpected qdiscs: %v", qdiscs)
	}

	netem, ok := qdiscs[0].(*netlink.Netem)
	if !ok {
		t.Fatalf("unexpected qdisc: %v", qdiscs[0])
	}

	if netem.Latency != 1562500 {
		t.Fatalf("unexpected latency: %v", netem.Latency)
	}

	if netem.Gap != 10 {
		t.Fatalf("unexpected gap: %v", netem.Gap)
	}
}

func TestParseQdiscsTbfNetem(t *testing.T) {
	intfProps := &qdisc.DeviceProperties{
		Rate:    "100 mbps",
		Latency: "100ms",
		Gap:     10,
		Loss:    10.0,
	}

	qdiscs, err := intfProps.ParseQdiscs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(qdiscs) != 2 {
		t.Fatalf("unexpected qdiscs: %v", qdiscs)
	}

}
