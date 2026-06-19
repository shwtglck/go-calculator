package calculator

import "fmt"

// Calculate выполняет математическую операцию над двумя числами.
func Calculate(a, b float64, operator string) (float64, error) {
	switch operator {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, fmt.Errorf("деление на ноль невозможно")
		}
		return a / b, nil
	default:
		return 0, fmt.Errorf("неизвестный оператор %q, используйте +, -, *, /", operator)
	}
}
