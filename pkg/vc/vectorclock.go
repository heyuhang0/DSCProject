package vc

// NewVectorClock
// MergeClock
// Advance
// ToDTO
// FromDTO

type vclock map[int]int // vector clock type

func NewVectorClock(machineID int) vclock {
	return vclock{machineID: 0}
}

func (selfclock vclock) MergeClock(machineID int, otherclock vclock) bool {
	// Be careful that map has no order
	result := false
	numOfEqual := 0
	// merge first
	for k, v := range otherclock {
		if val, ok := selfclock[k]; ok {
			// If the local clock is after (in line with the vector comparison)
			// the message clock, then a (potential) causality violation is flagged
			if val > v {
				result = true
			} else if val == v {
				// After means more than or equal, >=,
				// but CANT be the case where all elements are equal
				numOfEqual++
			} else {
				selfclock[k] = v
			}
		} else {
			// this key from otherclock is not in selfclock
			selfclock[k] = v
		}
	}
	// advance clock now
	selfclock.Advance(machineID)
	// need to return causality violation
	if numOfEqual == len(selfclock) {
		result = true
	}
	return result
}

func (selfclock vclock) Advance(machineID int) {
	selfclock[machineID]++
}

func ToDTO() {

}

func FromDTO() {

}
