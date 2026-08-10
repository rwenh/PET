package main

import (
	"strings"
	"time"
)

// Summary holds aggregate statistics over a set of expenses.
type Summary struct {
	Total   float64
	Count   int
	Average float64
	Max     float64
	Min     float64
}

func FilterByMonth(expenses []Expense, month, year int) []Expense {
	if year == 0 {
		year = time.Now().Year()
	}
	var result []Expense
	for _, e := range expenses {
		d, err := time.Parse(dateLayout, e.Date)
		if err != nil {
			continue
		}
		if int(d.Month()) == month && d.Year() == year {
			result = append(result, e)
		}
	}
	return result
}

func FilterByCategory(expenses []Expense, category string) []Expense {
	cat := strings.ToLower(category)
	var result []Expense
	for _, e := range expenses {
		if strings.ToLower(e.Category) == cat {
			result = append(result, e)
		}
	}
	return result
}

func CalculateSummary(expenses []Expense) Summary {
	if len(expenses) == 0 {
		return Summary{}
	}
	s := Summary{
		Count: len(expenses),
		Max:   expenses[0].Amount,
		Min:   expenses[0].Amount,
	}
	for _, e := range expenses {
		s.Total += e.Amount
		if e.Amount > s.Max {
			s.Max = e.Amount
		}
		if e.Amount < s.Min {
			s.Min = e.Amount
		}
	}
	s.Average = s.Total / float64(s.Count)
	return s
}

func GroupByCategory(expenses []Expense) map[string]float64 {
	m := make(map[string]float64)
	for _, e := range expenses {
		m[e.Category] += e.Amount
	}
	return m
}
