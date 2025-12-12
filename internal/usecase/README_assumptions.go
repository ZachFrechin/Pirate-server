package usecase

// Assumptions
//
// The game design provided does not specify deck size/composition.
// The server therefore defines a scalable deck in domain.buildDeck:
// per player: 6x(+1), 4x(0), 3x(-2), 3x(accusation).
// This keeps score/accusation reasonably balanced and long enough to play.
//
// If you want an exact deck design later (e.g. fixed total card counts),
// replace domain.buildDeck and keep the rest unchanged.
