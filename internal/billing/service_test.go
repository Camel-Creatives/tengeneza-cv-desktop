package billing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlansCatalogIncludesFreeAndPurchasablePacks(t *testing.T) {
	plans := Plans()
	require.NotEmpty(t, plans)
	free, ok := PlanByCode("free")
	require.True(t, ok)
	require.Equal(t, int64(0), free.AmountTZS)
	pack, ok := PlanByCode("starter_5")
	require.True(t, ok)
	require.Equal(t, 5, pack.ExportCredits)
	_, ok = PlanByCode("does_not_exist")
	require.False(t, ok)
}
