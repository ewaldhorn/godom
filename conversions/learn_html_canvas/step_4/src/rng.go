package main

// For the trees, I wanted each tree to look random, but I wanted that randomness to be tied to
// the tree height. That way, I can have trees render at any xy position, but their look is
// determined by their height. So it's sort of deterministic.
//
// I used https://rosettacode.org/wiki/Pseudo-random_numbers/Splitmix64 as a reference for building
// this. It's something I've been using since my Java days!

// ------------------------------------------------------------------------------------------------
type rng struct{ s uint64 }

// ------------------------------------------------------------------------------------------------
func newRNG(seed uint64) *rng {
	return &rng{s: seed}
}

// ------------------------------------------------------------------------------------------------
// next uses the golden ration (0x9E3779B97F4A7C15) to advance the seed and generated the next
// random number. Basically, this is SplitMix64, which is pretty cool for games. I tend to use
// it quite often because it's lightweight but efficient. Don't build your crypto exchange on it
// though, unless you want to get famous for the wrong reasons!
func (r *rng) next() uint64 {
	r.s += 0x9E3779B97F4A7C15
	z := r.s
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// ------------------------------------------------------------------------------------------------
func (r *rng) float64() float64 {
	return float64(r.next()>>11) / float64(1<<53)
}

// ------------------------------------------------------------------------------------------------
func (r *rng) intn(n int) int {
	return int(r.next() % uint64(n))
}
