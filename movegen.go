package movegengo

// MoveGen chess move generator
type MoveGen struct {
	moves  [255]uint16
	index  uint
	state  *GameState
	mover  *Move
	colour uint8
}

// NewMoveGen initiate a new MoveGen instance
func NewMoveGen() *MoveGen {
	return &MoveGen{index: 0, state: NewGameState(), mover: NewMove(0), colour: DefaultGameStateColour()}
}

// NewMoveGenByState creates a new movegen instance using the following state
func NewMoveGenByState(st *GameState) *MoveGen {
	return &MoveGen{index: 0, state: st, mover: NewMove(0), colour: (st.info & 0x10) >> 5}
}

// SetState update gamestate by reusing memory space
func (mg *MoveGen) SetState(st *GameState) {
	mg.state = st

	// reset any data associated with the old state
	mg.Clear()
}

// Clear the moves list tracker & iterator
func (mg *MoveGen) Clear() {
	mg.index = 0
}

// Size get the number of stored moves so far
func (mg *MoveGen) Size() uint {
	return mg.index
}

// AddMove to the stack
func (mg *MoveGen) AddMove(move uint16) {
	mg.moves[mg.index] = move
	mg.index++
}

// SetMove set a move at a specific location
func (mg *MoveGen) SetMove(move uint16, index int) {
	mg.moves[index] = move
}

// GetMove returns the move for a given index
func (mg *MoveGen) GetMove(index uint) uint16 {
	return mg.moves[index]
}

// CreateIterator Iterator pattern
func (mg *MoveGen) CreateIterator() *MoveGenIterator {
	return NewMoveGenIterator(0, mg.index, mg)
}

func (mg *MoveGen) isWhite() bool {
	return mg.colour == 1
}

// Moves are generated here
//

// GenerateMoves generates all the moves for an active player
func (mg *MoveGen) GenerateMoves() {
	if mg.isWhite() { // if the active player is white

	} else { // generate moves for the black player

	}

	mg.GeneratePawnMoves()
	mg.GenerateKnightMoves()
}

func (mg *MoveGen) GeneratePawnMoves() uint64 {
	pawns := mg.state.pieces[mg.colour*6]

	// attack left
	attacksLeft := mg.generatePawnLeftAttack(pawns)

	// attack right
	attacksRight := mg.generatePawnRightAttack(pawns)

	// single push
	singlePush := mg.generatePawnSinglePush()

	// double push
	mg.generatePawnDoublePush(singlePush)

	// results in possible attacks
	return attacksLeft | attacksRight
}

// generatePawnSinglePush Move all the pawns forward once and do a promotion check.
// Promotions are handled, and removed from the resulting bitboard.
//
// return A bitboard for all the pawns that moved forward, without being promoted.
func (mg *MoveGen) generatePawnSinglePush() uint64 {
	pawns := mg.state.pieces[mg.colour*6]
	var to uint64

	if mg.isWhite() {
		to = (pawns << 8) &^ mg.state.colours[0] // ~(this->state.taken)
	} else {
		to = (pawns >> 8) &^ mg.state.colours[1] // ~(this->state.taken)
	}
	cache := to

	// remove promotion pieces
	to ^= mg.generatePromotions(0, to)

	var attackDirection int
	if mg.isWhite() {
		attackDirection = -8
	} else {
		attackDirection = 8
	}

	// single push
	mg.mover.SetFlags(0)
	for i := LSB(to); i != 64; i = NLSB(&to, i) {
		mg.mover.SetFrom(uint16(i + attackDirection))
		mg.mover.SetTo(uint16(i))
		mg.moves[mg.index] = mg.mover.GetMove()
		mg.index++
	}

	return cache
} // end pawn generation

// generatePawnDoublePush Generate legal double push pawn moves.
// This is a continuation on single pawn push, so the argument must be
// the pawns that has already moved once for accurate results.
//
// param `pawns` All pawn positions after single legal push (move). eg. 16711680ull
//
// return All the new pawn positions
func (mg *MoveGen) generatePawnDoublePush(pawns uint64) uint64 {
	var to uint64

	if mg.isWhite() {
		to = ((pawns & 0xff0000) << 8) &^ mg.state.colours[0] // ~(this->state.taken)
	} else {
		to = ((pawns & 0xff0000000000) >> 8) &^ mg.state.colours[1] // ~(this->state.taken)
	}

	var attackDirection int
	if mg.isWhite() {
		attackDirection = -16
	} else {
		attackDirection = 16
	}

	cache := to
	mg.mover.SetFlags(1) // 0b0001, double push
	for i := LSB(to); i != 64; i = NLSB(&to, i) {
		mg.mover.SetFrom(uint16(i + attackDirection))
		mg.mover.SetTo(uint16(i))
		mg.moves[mg.index] = mg.mover.GetMove()
		mg.index++
	}

	return cache
}

// generatePawnLeftAttack Generate all pawn attacks on the left side. Promotions are handled as after movement.
//
// param pawns uint64 over all pawns
//
// return uint64 with positions reached
func (mg *MoveGen) generatePawnLeftAttack(pawns uint64) uint64 {
	area := uint64(0x7f7f7f7f7f7f7f7f)
	var attacks uint64
	var attackDirection int
	if mg.isWhite() {
		attackDirection = 9 // promotions
		attacks = (pawns & area) << 9
		attacks &= mg.state.colours[0] // TODO: en passant
	} else {
		attackDirection = -7 // promotions
		attacks = (pawns & area) >> 7
		attacks &= mg.state.colours[1] // TODO: en passant
	}

	cache := attacks

	// promotions
	attacks ^= mg.generatePromotions(attackDirection, attacks)

	mg.mover.SetFlags(4) // 0b0100, capture
	for i := LSB(attacks); i != 64; i = NLSB(&attacks, i) {
		mg.mover.SetFrom(uint16(i + attackDirection))
		mg.mover.SetTo(uint16(i))
		mg.moves[mg.index] = mg.mover.GetMove()
		mg.index++
	}

	return cache
}

// generatePawnRightAttack Generate all pawn attacks on the right side. Promotions are handled as after movement.
//
// param pawns uint64 over all pawns
//
// return uint64 with positions reached.
func (mg *MoveGen) generatePawnRightAttack(pawns uint64) uint64 {
	area := uint64(0xfefefefefefe00)
	var attacks uint64 // TODO: en passant
	var attackDirection int
	if mg.isWhite() {
		attackDirection = -7 // promotions
		attacks = (pawns & area) << 7
		attacks &= mg.state.colours[0]
	} else {
		attackDirection = 9 // promotions
		attacks = (pawns & area) >> 9
		attacks &= mg.state.colours[1]
	}
	cache := attacks

	// promotions
	attacks ^= mg.generatePromotions(attackDirection, attacks)

	// capture, 0b0100
	mg.mover.SetFlags(4)
	for i := LSB(attacks); i != 64; i = NLSB(&attacks, i) {
		mg.mover.SetFrom(uint16(i + attackDirection))
		mg.mover.SetTo(uint16(i))
		mg.moves[mg.index] = mg.mover.GetMove()
		mg.index++
	}

	return cache
}

// generatePromotions Generate promotion pieces.
//
// param FROM the bitboard offset in int8, diff between from and to position index.
func (mg *MoveGen) generatePromotions(from int, pawns uint64) uint64 {
	promotions := pawns & 0xff000000000000ff

	// single push has a FROM of 1. since this is an offset.
	var flag uint8
	if from != 8 && from != -8 {
		flag = 12 // 0b1100
	} else {
		flag = 8 // 0b1000
	}

	for i := LSB(promotions); i != 64; i = NLSB(&promotions, i) {
		mg.mover.SetFrom(uint16(i + from))
		mg.mover.SetTo(uint16(i))

		var t uint8
		for ; t < 4; t++ {
			mg.mover.SetFlags(uint16(flag + t))
			mg.moves[mg.index] = mg.mover.GetMove()
			mg.index++
		}
	}

	return pawns & 0xff000000000000ff
}

func (mg *MoveGen) generateKnightBoard(index int) uint64 {

	//attacks := uint64(0x28440044280000)
	//var result uint64

	//if (x & 0x3c3c3c3c0000) > 0 {
	//	origin := 45
	//	offset := uint16(origin - LSB(x))

	//	result = attacks >> offset
	//} else {
	//mask := uint64(0x7c7c7c7c7c0000)
	//origin := uint64(0x1000000000)
	//}

	return KnightMoves[index]
}

// GenerateKnightMoves ...
func (mg *MoveGen) GenerateKnightMoves() uint64 {
	knights := mg.state.pieces[mg.colour*6+2]
	attacks := uint64(0)
	for i := LSB(knights); i != 64; i = NLSB(&knights, i) {
		moves := mg.generateKnightBoard(i)
		attacks |= moves

		var moveAttacks uint64
		if mg.isWhite() {
			moveAttacks = moves & mg.state.colours[0]
		} else {
			moveAttacks = moves & mg.state.colours[1]
		}
		moves ^= moveAttacks
		moves &= ^mg.state.colours[mg.colour]

		mg.mover.SetFrom(uint16(i))
		mg.mover.SetFlags(0) // quiet move
		for j := LSB(moves); j != 64; j = NLSB(&moves, j) {
			mg.mover.SetTo(uint16(j))
			mg.moves[mg.index] = mg.mover.GetMove()
			mg.index++
		}

		mg.mover.SetFlags(8) // captures
		for j := LSB(moveAttacks); j != 64; j = NLSB(&moveAttacks, j) {
			mg.mover.SetTo(uint16(j))
			mg.moves[mg.index] = mg.mover.GetMove()
			mg.index++
		}

	}

	return attacks
}
