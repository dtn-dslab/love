package intf

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
)

type DeviceProperties struct {
	Rate          Rate
	Latency       Duration
	DelayCorr     Percentage
	Jitter        Duration
	Loss          Percentage
	LossCorr      Percentage
	Gap           uint32
	Duplicate     Percentage
	DuplicateCorr Percentage
	ReorderProb   Percentage
	ReorderCorr   Percentage
	CorruptProb   Percentage
	CorruptCorr   Percentage
}

func (p *DeviceProperties) ParseQdiscs() ([]netlink.Qdisc, error) {
	qdiscs := make([]netlink.Qdisc, 0)

	rate, err := p.Rate.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse rate: %w", err)
		}
	} else {
		qdiscs = append(qdiscs, &netlink.Tbf{
			Rate:     rate,
			Buffer:   getTbfBurst(rate),
			Minburst: 1500,
		})
	}

	enableNetem := false

	latency, err := p.Latency.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse latency: %w", err)
		}
	} else {
		enableNetem = true
	}

	delayCorr, err := p.DelayCorr.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse delay correlation: %w", err)
		}
	} else {
		enableNetem = true
	}

	jitter, err := p.Jitter.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse jitter: %w", err)
		}
	} else {
		enableNetem = true
	}

	loss, err := p.Loss.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse loss: %w", err)
		}
	} else {
		enableNetem = true
	}

	lossCorr, err := p.LossCorr.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse loss correlation: %w", err)
		}
	} else {
		enableNetem = true
	}

	duplicate, err := p.Duplicate.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse duplicate: %w", err)
		}
	} else {
		enableNetem = true
	}

	duplicateCorr, err := p.DuplicateCorr.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse duplicate correlation: %w", err)
		}
	} else {
		enableNetem = true
	}

	reorderProb, err := p.ReorderProb.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse reorder probability: %w", err)
		}
	} else {
		enableNetem = true
	}

	reorderCorr, err := p.ReorderCorr.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse reorder correlation: %w", err)
		}
	} else {
		enableNetem = true
	}

	corruptProb, err := p.CorruptProb.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse corrupt probability: %w", err)
		}
	} else {
		enableNetem = true
	}

	corruptCorr, err := p.CorruptCorr.Parse()
	if err != nil {
		if _, ok := err.(*EmptyValueError); !ok {
			return nil, fmt.Errorf("failed to parse corrupt correlation: %w", err)
		}
	} else {
		enableNetem = true
	}

	if enableNetem {
		qdiscs = append(qdiscs, netlink.NewNetem(netlink.QdiscAttrs{}, netlink.NetemQdiscAttrs{
			Latency:       latency,
			DelayCorr:     delayCorr,
			Jitter:        jitter,
			Loss:          loss,
			LossCorr:      lossCorr,
			Gap:           p.Gap,
			Duplicate:     duplicate,
			DuplicateCorr: duplicateCorr,
			ReorderProb:   reorderProb,
			ReorderCorr:   reorderCorr,
			CorruptProb:   corruptProb,
			CorruptCorr:   corruptCorr,
		}))
	}

	return qdiscs, nil
}

// Bandwidth rate limit, e.g. 1000(bit/s), 100kbit, 100Mbps, 1Gibps.
// For more information, refer to https://man7.org/linux/man-pages/man8/tc.8.html.
type Rate string

func (r Rate) Parse() (uint64, error) {
	rate := strings.TrimSpace(strings.ToLower(string(r)))
	if rate == "" {
		return 0, &EmptyValueError{}
	}

	var unitMultiplier uint64 = 1
	if strings.HasSuffix(rate, "bit") {
		rate = strings.TrimSuffix(rate, "bit")
	} else if strings.HasSuffix(rate, "bps") {
		rate = strings.TrimSuffix(rate, "bps")
		unitMultiplier = 8
	}

	// Assume SI-prefixes by default
	var base uint64 = 1000
	// If using IEC-prefixes, switch to binary base, e.g. MiB
	if strings.HasSuffix(rate, "i") {
		rate = strings.TrimSuffix(rate, "i")
		base = 1024
	}

	for i, unit := range []string{"k", "m", "g", "t"} {
		if strings.HasSuffix(rate, unit) {
			rate = strings.TrimSuffix(rate, unit)
			for j := 0; j < i+1; j++ {
				unitMultiplier *= base
			}
			break
		}
	}

	rate = strings.TrimSpace(rate)
	value, err := strconv.ParseUint(rate, 10, 64)
	if err != nil {
		return 0, err
	}
	return value * unitMultiplier, nil
}

// Duration string format, e.g. "300ms", "1.5s".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
type Duration string

func (d Duration) Parse() (uint32, error) {
	if d == "" {
		return 0, &EmptyValueError{}
	}
	value, err := time.ParseDuration(string(d))
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, fmt.Errorf("duration value must be positive")
	}
	return uint32(value.Microseconds()), nil
}

// Percentage is a float32 value between 0 and 100.
type Percentage float32

func (p Percentage) Parse() (float32, error) {
	if p == 0 {
		return 0, &EmptyValueError{}
	} else if p < 0 || p > 100 {
		return 0, fmt.Errorf("percentage value must be between 0 and 100")
	} else {
		return float32(p), nil
	}
}

type EmptyValueError struct{}

func (e EmptyValueError) Error() string {
	return "value is empty"
}

// Calculate burst size for TBF qdisc
func getTbfBurst(rate uint64) uint32 {
	// At least Rate / Kernel HZ
	burst := uint32(rate / 250)
	// At least 5000 bytes
	if burst < 5000 {
		burst = 5000
	}
	return burst
}
