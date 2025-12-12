package domain

// CardType distinguishes playable cards.
type CardType string

const (
	CardTypeScore      CardType = "score"
	CardTypeAccusation CardType = "accusation"
)

// Card is a single card in the deck/hand.
//
// For CardTypeScore, Score is one of: +1, 0, -2.
// For CardTypeAccusation, Score is always 0.
type Card struct {
	Type  CardType `json:"type"`
	Score int      `json:"score"`
}

func (c Card) IsScore() bool {
	return c.Type == CardTypeScore
}

func (c Card) IsAccusation() bool {
	return c.Type == CardTypeAccusation
}
