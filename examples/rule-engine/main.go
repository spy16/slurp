package main

import (
	"log"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/core"
)

func main() {
	// Accept business rules from file, command-line, http request etc.
	// These rules can change as per business requirements and your
	// application doesn't have to change.
	ruleSrc := `(and (regular-user? current-user)
					 (not-blacklisted? current-user))`

	shouldDiscount, err := runDiscountingRule(ruleSrc, "bob")
	if err != nil {
		panic(err)
	}

	if shouldDiscount {
		// apply discount for the order
		log.Printf("applying discount")
	} else {
		// don't apply discount
		log.Printf("not applying discount")
	}
}

func runDiscountingRule(rule string, user string) (bool, error) {
	// Define and expose your rules which ideally should have no
	// side effects.
	globals := map[string]core.Any{
		"and":                 slurp.Func("and", and),
		"or":                  slurp.Func("or", or),
		"regular-user?":       slurp.Func("isRegularUser", isRegularUser),
		"minimum-cart-price?": slurp.Func("isMinCartPrice", isMinCartPrice),
		"not-blacklisted?":    slurp.Func("isNotBlacklisted", isNotBlacklisted),
		"current-user":        user,
	}

	ins := slurp.New()
	_ = ins.Bind(globals)
	shouldDiscount, err := ins.EvalStr(rule)
	return core.IsTruthy(shouldDiscount), err
}

func isNotBlacklisted(user string) bool {
	return user != "joe"
}

func isMinCartPrice(price float64) bool {
	return price >= 100
}

func isRegularUser(user string) bool {
	return user == "bob"
}

func and(rest ...bool) bool {
	if len(rest) == 0 {
		return true
	}
	result := rest[0]
	for _, r := range rest {
		result = result && r
		if !result {
			return false
		}
	}
	return true
}

func or(rest ...bool) bool {
	if len(rest) == 0 {
		return true
	}
	result := rest[0]
	for _, r := range rest {
		if result {
			return true
		}
		result = result || r
	}
	return false
}
