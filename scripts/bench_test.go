package scripts

import (
	"strings"
	"testing"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/routing"
	"github.com/dresar/ekarouter/internal/tokensaver"
)

func BenchmarkTokenSaverCompact(b *testing.B) {
	ts := tokensaver.New("safe")
	diff := "diff --git a/file.go b/file.go\n" + strings.Repeat("+ added content line\n", 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ts.Compact(diff)
	}
}

func BenchmarkTokenSaverRepeatedLines(b *testing.B) {
	input := strings.Repeat("error line\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokensaver.CompactRepeatedLines(input)
	}
}

func BenchmarkAuthHashToken(b *testing.B) {
	token := "eka_live_abcdef1234567890abcdef1234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = auth.HashToken(token)
	}
}

func BenchmarkRouterSelectTargets(b *testing.B) {
	cd := routing.NewCooldownManager()
	r := routing.NewRouter(cd)
	r.SetAccount(&routing.Account{ID: "acc-1", State: "active", Enabled: true})
	r.SetRoute(&routing.Route{
		Name:     "gpt-4o",
		Strategy: routing.StrategyPriority,
		Enabled:  true,
		Items: []routing.RouteItem{
			{ProviderID: "p1", ProviderKind: "openai", AccountID: "acc-1", ModelName: "gpt-4o", Priority: 10, Enabled: true},
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.SelectTargets("gpt-4o")
	}
}
