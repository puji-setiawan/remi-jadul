package game

// Additional jago/pickup aliases.
func (r *Round) TakeJago(player int, args ...interface{}) error   { return pickupImpl(r, player, args) }
func (r *Round) JagoPickup(player int, args ...interface{}) error { return pickupImpl(r, player, args) }
func (r *Round) PickupJago(player int, args ...interface{}) error { return pickupImpl(r, player, args) }
func (r *Round) TakeDiscardCardWithJago(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}
func (r *Round) TakeDiscardWithJago(player int, args ...interface{}) error {
	return pickupImpl(r, player, args)
}

// Additional laydown aliases.
func (r *Round) LayCards(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) LayCombinationCards(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) LayDownCards(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) PutCombination(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) LockHandCombination(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}

// Additional close aliases.
func (r *Round) CloseGame(player int, args ...interface{}) error { return closeImpl(r, player, args) }
func (r *Round) CompleteClose(player int, args ...interface{}) error {
	return closeImpl(r, player, args)
}
func (r *Round) FinishClose(player int, args ...interface{}) error { return closeImpl(r, player, args) }
func (r *Round) EndClose(player int, args ...interface{}) error    { return closeImpl(r, player, args) }
func (r *Round) CloseBy(player int, args ...interface{}) error     { return closeImpl(r, player, args) }

// Additional aliases for ended state.
func (r *Round) MarkClosed(player int) {
	r.Closed = true
	r.HasClosed = true
	r.ClosePlayer = player
	r.Closer = player
	r.ClosedBy = player
}

func (r *Round) Laydown(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) LayDownCombination(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) LaydownCombination(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
func (r *Round) LayCombinationDown(player int, args ...interface{}) error {
	return layCombinationImpl(r, player, args)
}
