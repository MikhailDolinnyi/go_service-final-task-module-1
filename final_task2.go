package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"unicode"
)

// Приоритет операторов
var priority = map[rune]int{
	'+': 1,
	'-': 1,
	'*': 2,
	'/': 2,
}

func Operations(op rune, a, b float64) (float64, error) {
	switch op {
	case '+':
		return a + b, nil
	case '-':
		return a - b, nil
	case '*':
		return a * b, nil
	case '/':
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	}
	return 0, errors.New("unknown operator")
}

func OperationsProcessing(op rune, stack *[]float64, operators *[]rune) error {
	for len(*operators) > 0 && (*operators)[len(*operators)-1] != '(' && priority[(*operators)[len(*operators)-1]] >= priority[op] {
		topOp := (*operators)[len(*operators)-1]
		*operators = (*operators)[:len(*operators)-1]

		if len(*stack) < 2 {
			return errors.New("invalid expression")
		}

		b := (*stack)[len(*stack)-1]
		a := (*stack)[len(*stack)-2]
		*stack = (*stack)[:len(*stack)-2]

		res, err := Operations(topOp, a, b)
		if err != nil {
			return err
		}

		*stack = append(*stack, res)
	}
	*operators = append(*operators, op)
	return nil
}

func Calc(expression string) (float64, error) {
	var stack []float64
	var operators []rune

	for i := 0; i < len(expression); i++ {
		char := rune(expression[i])

		if unicode.IsSpace(char) {
			continue
		}

		if unicode.IsDigit(char) || char == '.' {
			start := i
			for i < len(expression) && (unicode.IsDigit(rune(expression[i])) || rune(expression[i]) == '.') {
				i++
			}
			num, err := strconv.ParseFloat(expression[start:i], 64)
			if err != nil {
				return 0, errors.New("invalid number")
			}
			stack = append(stack, num)
			i--
		} else if char == '(' {
			operators = append(operators, char)
		} else if char == ')' {
			for len(operators) > 0 && operators[len(operators)-1] != '(' {
				topOp := operators[len(operators)-1]
				operators = operators[:len(operators)-1]

				if len(stack) < 2 {
					return 0, errors.New("invalid expression")
				}

				b := stack[len(stack)-1]
				a := stack[len(stack)-2]
				stack = stack[:len(stack)-2]

				res, err := Operations(topOp, a, b)
				if err != nil {
					return 0, err
				}

				stack = append(stack, res)
			}
			if len(operators) == 0 {
				return 0, errors.New("mismatched parentheses")
			}
			operators = operators[:len(operators)-1]
		} else if char == '+' || char == '-' || char == '*' || char == '/' {
			if err := OperationsProcessing(char, &stack, &operators); err != nil {
				return 0, err
			}
		} else {
			return 0, errors.New("invalid character in expression")
		}
	}

	for len(operators) > 0 {
		topOp := operators[len(operators)-1]
		operators = operators[:len(operators)-1]

		if len(stack) < 2 {
			return 0, errors.New("invalid expression")
		}

		b := stack[len(stack)-1]
		a := stack[len(stack)-2]
		stack = stack[:len(stack)-2]

		res, err := Operations(topOp, a, b)
		if err != nil {
			return 0, err
		}

		stack = append(stack, res)
	}

	if len(stack) != 1 {
		return 0, errors.New("invalid expression")
	}

	return stack[0], nil
}

// HTTP обработчик для вычисления выражений
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	result, err := Calc(reqBody.Expression)
	if err != nil {
		if err.Error() == "invalid character in expression" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{"error": "Expression is not valid"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"result": result})
}

func main() {
	http.HandleFunc("/api/v1/calculate", calculateHandler)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
