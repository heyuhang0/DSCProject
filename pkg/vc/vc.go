package vc

import (
	"fmt"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"sync"
)

type VectorClock struct {
	Vclock    map[int]int // vector clock type
	machineID int
	mu        sync.RWMutex
}

func NewVectorClock(machineID int) *VectorClock {
	return &VectorClock{
		Vclock:    map[int]int{machineID: 0},
		machineID: machineID,
		mu:        sync.RWMutex{},
	}
}

func (selfVectorClock *VectorClock) String() string {
	if selfVectorClock == nil {
		return "<nil>"
	}
	return fmt.Sprintf("&{%v %v}", selfVectorClock.machineID, selfVectorClock.Vclock)
}

func (selfVectorClock *VectorClock) MergeClock(otherClock map[int]int) bool {
	// Be careful that map has no order
	result := false
	numOfEqual := 0
	selfclock := selfVectorClock.Vclock
	// merge first
	selfVectorClock.mu.Lock()
	for k, v := range otherClock {
		if val, ok := selfclock[k]; ok {
			// If the local clock is after (in line with the vector comparison)
			// the message clock, then a (potential) causality violation is flagged
			if val > v {
				result = true
			} else if val == v {
				// After means more than or equal, >=
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
	// Release read lock
	selfVectorClock.mu.Unlock()
	// advance clock now
	selfVectorClock.Advance()
	// need to return causality violation
	if numOfEqual == len(selfclock) {
		result = true
	}
	return result
}

func (selfVectorClock *VectorClock) Advance() {
	selfVectorClock.mu.Lock()
	selfVectorClock.Vclock[selfVectorClock.machineID]++
	selfVectorClock.mu.Unlock()
}

func (selfVectorClock *VectorClock) GetVectorClock() map[int]int {
	selfVectorClock.mu.RLock()
	c := make(map[int]int)
	for k, v := range selfVectorClock.Vclock {
		c[k] = v
	}
	selfVectorClock.mu.RUnlock()
	return c
}

func ToDTO(selfVectorClock *VectorClock) *pb.VectorClock {
	result := make(map[int64]int64)
	selfVectorClock.mu.RLock()
	for k, v := range selfVectorClock.Vclock {
		result[int64(k)] = int64(v)
	}
	selfVectorClock.mu.RUnlock()
	return &pb.VectorClock{Vclock: result, MachineID: int64(selfVectorClock.machineID)}
}

func FromDTO(vc *pb.VectorClock) *VectorClock {
	data := vc.Vclock
	result := make(map[int]int)
	for k, v := range data {
		result[int(k)] = int(v)
	}
	return &VectorClock{
		Vclock:    result,
		machineID: int(vc.MachineID),
	}
}
