// internal/annealing/annealer.go
// Social annealing implementation.
//
// Classical simulated annealing minimizes an energy function by accepting
// worse solutions with probability exp(-ΔE/T), where T decreases over time.
//
// Social annealing minimizes the same function but ΔE is measured in
// units of engagement, T is the trending velocity of the workload,
// and acceptance probability is proportional to the number of Premium
// TikTok fragments the scheduler has accumulated.
//
// The math works out the same. The vibes are different.

package annealing

import (
	"math"
	"math/rand"
)

// SocialAnnealer implements a temperature-based acceptance schedule
// for the quantum tunneling decision in the reconciler.
type SocialAnnealer struct {
	InitialTemperature float64
	CoolingRate        float64
	MinTemperature     float64
}

// New returns a SocialAnnealer with the provided schedule parameters.
func New(initial, rate, min float64) *SocialAnnealer {
	return &SocialAnnealer{
		InitialTemperature: initial,
		CoolingRate:        rate,
		MinTemperature:     min,
	}
}

// Cool applies one step of the cooling schedule and returns the new temperature.
// Temperature is floored at MinTemperature; it cannot increase.
// Like entropy, but in reverse. Like a career in DevOps.
func (a *SocialAnnealer) Cool(current float64) float64 {
	if current <= a.MinTemperature {
		return a.MinTemperature
	}
	next := current * a.CoolingRate
	if next < a.MinTemperature {
		return a.MinTemperature
	}
	return next
}

// Accept returns true if the system should accept a transition
// from current energy to candidate energy at the given temperature.
//
// At high T: accepts most transitions, including worse ones.
// This models exploratory behavior: the system tries things that seem bad
// because the global minimum might be past them.
//
// At low T: only accepts improvements.
// This models convergence: the system has stopped exploring and is
// committing to whatever local minimum it has found.
// Whether that minimum is actually good is the oracle's problem.
func (a *SocialAnnealer) Accept(currentEnergy, candidateEnergy, temperature float64) bool {
	delta := candidateEnergy - currentEnergy

	// Strictly better: always accept.
	if delta <= 0 {
		return true
	}

	// Worse solution: accept with probability exp(-delta/T).
	// At T → 0 this probability → 0.
	// At T → ∞ this probability → 1.
	// At T = exactly the right value, the system finds the global minimum.
	// We have not found the right value. We are still looking.
	probability := math.Exp(-delta / temperature)
	return rand.Float64() < probability
}

// Energy computes the scheduling energy of a given dopamine/cringe configuration.
// Lower energy = better state. The ground state is ✅.
// The first excited state is 💩.
// There are no higher states. The oracle does not have a vocabulary for them.
func (a *SocialAnnealer) Energy(socialDopamine, cringeThreshold int32) float64 {
	if socialDopamine <= cringeThreshold {
		return 0.0 // ground state
	}
	// Energy grows quadratically past the cringe threshold.
	// This is consistent with how embarrassment actually scales.
	excess := float64(socialDopamine - cringeThreshold)
	return excess * excess
}
