package planner

import "fmt"

const unlimitedTimes = uint(0)

func recovered(v any) string {
	switch v := v.(type) {
	case error:
		return v.Error()

	case string:
		return v
	}

	return fmt.Sprintf("%+v", v)
}

func trackRepeatable(r Expectation) bool {
	t := r.RemainTimes()

	if t == unlimitedTimes {
		return true
	}

	return t > 1
}
