package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"pet_adopter/src/config"
)

type Color struct {
	R uint8
	G uint8
	B uint8
}

func ParseColor(color string) (Color, error) {
	parts := strings.Split(color, " ")
	if len(parts) != 3 {
		return Color{}, fmt.Errorf("len(parts) != 3")
	}

	r, err := strconv.ParseUint(parts[0], 10, 8)
	if err != nil {
		return Color{}, errors.Wrap(err, "failed to parse r")
	}

	g, err := strconv.ParseUint(parts[1], 10, 8)
	if err != nil {
		return Color{}, errors.Wrap(err, "failed to parse g")
	}

	b, err := strconv.ParseUint(parts[1], 10, 8)
	if err != nil {
		return Color{}, errors.Wrap(err, "failed to parse b")
	}

	return Color{R: uint8(r), G: uint8(g), B: uint8(b)}, nil
}

func Distance(left Color, right Color, cfg config.ColorConfig) (int64, bool) {
	distR := abs(int64(left.R) - int64(right.R))
	distG := abs(int64(left.G) - int64(right.G))
	distB := abs(int64(left.B) - int64(right.B))
	distSum := distR + distG + distB

	return distSum, !(distR > cfg.MaxPartDistance || distG > cfg.MaxPartDistance || distB > cfg.MaxPartDistance || distSum > cfg.MaxSumDistance)
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
