package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBidStats(t *testing.T) {
	b := NewBidStats(3) // require 3 entries to start returning non-0

	// diff should be 0 if nothing there
	require.Equal(t, 0.0, b.PayloadSizeDeviation(100))

	b.AddEntry(0, 100)
	require.Equal(t, 0.0, b.PayloadSizeDeviation(100))

	b.AddEntry(0, 200)
	require.Equal(t, 0.0, b.PayloadSizeDeviation(100))

	// add third entry. avg will be 200
	b.AddEntry(0, 300)
	require.Equal(t, -0.5, b.PayloadSizeDeviation(100))

	// add fourth entry. avg will be (200+300+400)/3 = 300
	b.AddEntry(0, 400)
	require.Equal(t, 0.0, b.PayloadSizeDeviation(300))
	require.Equal(t, 3, len(b.entries))

	// now test the max entries
	b.AddEntry(0, 400)
	require.Equal(t, 3, len(b.entries))
}
