package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const Currency = "\u20B9"

func FormatTable(expenses []Expense, showIndex bool) {
	if len(expenses) == 0 {
		fmt.Println("No expenses found.")
		return
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))

	hdr := ""
	if showIndex {
		hdr = fmt.Sprintf("%-4s ", "#")
	}
	hdr += fmt.Sprintf("%-12s %12s %-15s Description", "Date", "Amount", "Category")
	fmt.Println(hdr)
	fmt.Println(strings.Repeat("-", 80))

	for i, e := range expenses {
		line := ""
		if showIndex {
			line = fmt.Sprintf("%-4d ", i)
		}
		line += fmt.Sprintf("%-12s %s%9.2f %-15s %s", e.Date, Currency, e.Amount, e.Category, e.Description)
		fmt.Println(line)
	}
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()
}

func FormatSummary(s Summary, expenses []Expense) {
	if s.Count == 0 {
		fmt.Println("No expenses to summarize.")
		return
	}
	c := Currency
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(centerString("EXPENSE SUMMARY", 60))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total:         %s%10s\n", c, formatCommas(s.Total))
	fmt.Printf("Count:          %11d\n", s.Count)
	fmt.Printf("Average:        %s%10s\n", c, formatCommas(s.Average))
	fmt.Printf("Highest:        %s%10s\n", c, formatCommas(s.Max))
	fmt.Printf("Lowest:         %s%10s\n", c, formatCommas(s.Min))

	if days, ok := spanDays(expenses); ok && days > 0 {
		daily := s.Total / float64(days)
		fmt.Printf("Daily average:  %s%10s\n", c, formatCommas(daily))
	}
	fmt.Println(strings.Repeat("=", 60))

	byCat := GroupByCategory(expenses)
	if len(byCat) > 0 && s.Total > 0 {
		fmt.Println("\nCategory breakdown:")
		fmt.Println(strings.Repeat("-", 60))
		for _, ca := range sortedByAmountDesc(byCat) {
			pct := ca.amount / s.Total * 100
			fmt.Printf("%-20s %s%11s  (%5.1f%%)\n", ca.category, c, formatCommas(ca.amount), pct)
		}
		fmt.Println(strings.Repeat("-", 60))
	}
}

func spanDays(expenses []Expense) (int, bool) {
	var min, max time.Time
	found := false
	for _, e := range expenses {
		d, err := time.Parse(dateLayout, e.Date)
		if err != nil {
			continue
		}
		if !found {
			min, max, found = d, d, true
			continue
		}
		if d.Before(min) {
			min = d
		}
		if d.After(max) {
			max = d
		}
	}
	if !found {
		return 0, false
	}
	return int(max.Sub(min).Hours()/24) + 1, true
}

type categoryAmount struct {
	category string
	amount   float64
}

func sortedByAmountDesc(byCat map[string]float64) []categoryAmount {
	list := make([]categoryAmount, 0, len(byCat))
	for cat, amt := range byCat {
		list = append(list, categoryAmount{cat, amt})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].amount > list[j].amount
	})
	return list
}

func centerString(s string, width int) string {
	if len(s) >= width {
		return s
	}
	left := (width - len(s)) / 2
	right := width - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func formatCommas(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	intPart, decPart, _ := strings.Cut(s, ".")

	neg := strings.HasPrefix(intPart, "-")
	if neg {
		intPart = intPart[1:]
	}

	n := len(intPart)
	var out strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteByte(intPart[i])
	}
	result := out.String() + "." + decPart
	if neg {
		result = "-" + result
	}
	return result
}
