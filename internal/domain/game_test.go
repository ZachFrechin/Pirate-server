package domain

import "testing"

func TestImpostorCount(t *testing.T) {
	cases := []struct {
		players int
		want    int
	}{
		{3, 1}, {4, 1}, {5, 1},
		{6, 2}, {7, 2}, {8, 2},
	}
	for _, tc := range cases {
		if got := impostorCount(tc.players); got != tc.want {
			t.Fatalf("players=%d got=%d want=%d", tc.players, got, tc.want)
		}
	}
}

func TestStartSetsGoalAndDealsHands(t *testing.T) {
	g := NewLobbyGame()
	for i := 0; i < 5; i++ {
		if err := g.AddPlayer(&Player{ID: string(rune('a' + i)), Name: "P"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(); err != nil {
		t.Fatal(err)
	}
	if g.Status != GameStatusInGame {
		t.Fatalf("status=%s", g.Status)
	}
	if g.GoalScore != 11 {
		t.Fatalf("goal=%d", g.GoalScore)
	}
	for _, p := range g.Players {
		if len(p.Hand) != StartingHandSize {
			t.Fatalf("hand=%d", len(p.Hand))
		}
	}
}

func TestAccusationEliminatesAtThree(t *testing.T) {
	g := NewLobbyGame()
	g.Players = []*Player{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B"},
		{ID: "c", Name: "C"},
	}
	_ = g.Start()

	// Force A's hand to be accusation cards.
	a, _ := g.mustPlayer("a")
	a.Hand = []Card{{Type: CardTypeAccusation}, {Type: CardTypeAccusation}, {Type: CardTypeAccusation}}
	g.TurnIndex = 0

	for i := 0; i < 3; i++ {
		if err := g.PlayAccusationCard("a", 0, "b"); err != nil {
			t.Fatal(err)
		}
		// Force turn back to A for this test.
		g.TurnIndex = 0
		// Put accusation card back.
		a.Hand = append(a.Hand, Card{Type: CardTypeAccusation})
	}

	b, _ := g.mustPlayer("b")
	if !b.Eliminated {
		t.Fatalf("expected eliminated")
	}
	if b.Accusations < AccusationsToEliminate {
		t.Fatalf("accusations=%d", b.Accusations)
	}
}
