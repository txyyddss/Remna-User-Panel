package database

type fixedActivityRandom struct {
	value int64
	err   error
}

type countingActivityRandom struct {
	value int64
	calls int
}

func (random *countingActivityRandom) Int63n(upperBound int64) (int64, error) {
	random.calls++
	return random.value % upperBound, nil
}

func (random fixedActivityRandom) Int63n(upperBound int64) (int64, error) {
	if random.err != nil {
		return 0, random.err
	}
	return random.value % upperBound, nil
}
