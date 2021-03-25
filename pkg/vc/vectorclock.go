package vc

import (
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
)

type vclock map[int]int // vector clock type

func NewVectorClock(machineID int) vclock {
	return vclock{machineID: 0, -1: machineID}
}

func (selfclock vclock) MergeClock(otherclock vclock) bool {
	// Be careful that map has no order
	result := false
	numOfEqual := 0
	// merge first
	for k, v := range otherclock {
		if k == -1 {
			// -1 key is the machineID
			continue
		}
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
	// advance clock now
	selfclock.Advance()
	// need to return causality violation
	if numOfEqual == len(selfclock) {
		result = true
	}
	return result
}

func (selfclock vclock) Advance() {
	machineID := selfclock[-1]
	selfclock[machineID]++
}

func ToDTO(data vclock) *pb.VectorClock {
	result := make(map[int64]int64)
	for k, v := range data {
		result[int64(k)] = int64(v)
	}
	return &pb.VectorClock{Vclock: result}
}

func FromDTO(vc *pb.VectorClock) vclock {
	data := vc.Vclock
	result := make(vclock)
	for k, v := range data {
		result[int(k)] = int(v)
	}
	return result
}
