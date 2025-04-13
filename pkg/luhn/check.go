package luhn

import "strconv"

func CheckString(s string) bool {
	number, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return Check(number)
}

func Check(number int) bool {
	var sum int
	for i := 0; number > 0; i++ {
		cur := number % 10
		if i%2 == 0 {
			sum += cur
			number /= 10
			continue
		}
		cur *= 2
		//nolint:mnd // ignore
		if cur > 9 {
			cur -= 9
		}
		sum += cur
		number /= 10
	}
	return sum%10 == 0
}
