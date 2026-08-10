// Command pet is a personal expense tracker with CSV storage
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const usage = `Personal Expense Tracker
Usage:
    pet <command> [arguments]
Commands:
    add Add a new expense
    list List expenses
    summary Show expense summary
    delete Delete an expense by index
Examples:
    pet add 420 "Street food" -c Food
    pet list --month 1
    pet summary --month 12 --year 2025
    pet delete 3
Run 'pet <command> -h' for command-specific help.`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}
	storage, err := NewExpenseStorage("expenses.csv")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch os.Args[1] {
	case "add":
		runAdd(storage, os.Args[2:])
	case "list":
		runList(storage, os.Args[2:])
	case "summary":
		runSummary(storage, os.Args[2:])
	case "delete":
		runDelete(storage, os.Args[2:])
	case "-h", "--help", "help":
		fmt.Println(usage)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		fmt.Fprintln(os.Stderr, usage)
	}
}

func validateAmount(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid positive amount: %s", s)
	}
	return v, nil
}

func validateDate(s string) error {
	if _, err := time.Parse(dateLayout, s); err != nil {
		return fmt.Errorf("invalid date %q, expected YYYY-MM-DD", s)
	}
	return nil
}

func looksLikeNegativeNumber(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	seenDigit, seenDot := false, false
	for _, r := range s[1:] {
		switch {
		case r >= '0' && r <= '9':
			seenDigit = true
		case r == '.' && !seenDot:
			seenDot = true
		default:
			return false
		}
	}
	return seenDigit
}

func splitFlagsAndPositionals(args []string, flagsWithValue map[string]bool) (flags, positionals []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") && !looksLikeNegativeNumber(a) {
			flags = append(flags, a)
			name := strings.TrimLeft(a, "-")
			if strings.Contains(name, "=") {
				continue // value is attached, e.g. --category=Food
			}
			if flagsWithValue[name] && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		positionals = append(positionals, a)
	}
	return flags, positionals
}

func runAdd(storage *ExpenseStorage, args []string) {
	flagArgs, positionals := splitFlagsAndPositionals(args, map[string]bool{
		"c": true, "category": true,
		"d": true, "date": true,
	})

	fs := flag.NewFlagSet("add", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage: pet add <amount> <description> [-c category] [-d date]`)
		fs.PrintDefaults()
	}
	category := fs.String("c", "General", "expense category")
	fs.StringVar(category, "category", "General", "expense category")
	date := fs.String("d", "", "expense date, YYYY-MM-DD (default: today)")
	fs.StringVar(date, "date", "", "expense date, YYYY-MM-DD (default: today)")
	fs.Parse(flagArgs)

	if len(positionals) != 2 {
		fmt.Fprintln(os.Stderr, "Error: add requires exactly <amount> and <description>")
		fs.Usage()
		os.Exit(2)
	}

	amount, err := validateAmount(positionals[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}

	if err := storage.SaveExpense(amount, *category, positionals[1], *date); err != nil {
		fmt.Println(" Failed:", err)
		return
	}
	fmt.Println(" Added")
}

func runList(storage *ExpenseStorage, args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	month := fs.Int("month", 0, "filter by month (1-12)")
	year := fs.Int("year", 0, "filter by year (default: current year)")
	category := fs.String("c", "", "filter by category")
	fs.StringVar(category, "category", "", "filter by category")
	showIndex := fs.Bool("show-index", false, "show each expense's index (needed for delete)")
	fs.Parse(args)

	expenses, err := storage.LoadExpenses()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *month != 0 {
		expenses = FilterByMonth(expenses, *month, *year)
	}
	if *category != "" {
		expenses = FilterByCategory(expenses, *category)
	}
	FormatTable(expenses, *showIndex)
	if len(expenses) > 0 {
		total := 0.0
		for _, e := range expenses {
			total += e.Amount
		}
		fmt.Printf("Total: %s%s (%d items)\n", Currency, formatCommas(total), len(expenses))
	}
}

func runSummary(storage *ExpenseStorage, args []string) {
	fs := flag.NewFlagSet("summary", flag.ExitOnError)
	month := fs.Int("month", 0, "filter by month (1-12)")
	year := fs.Int("year", 0, "filter by year (default: current year)")
	category := fs.String("c", "", "filter by category")
	fs.StringVar(category, "category", "", "filter by category")
	fs.Parse(args)

	expenses, err := storage.LoadExpenses()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *month != 0 {
		expenses = FilterByMonth(expenses, *month, *year)
	}
	if *category != "" {
		expenses = FilterByCategory(expenses, *category)
	}
	FormatSummary(CalculateSummary(expenses), expenses)
}

func runDelete(storage *ExpenseStorage, args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	fs.Parse(args)
	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "Error: delete requires <index>")
		os.Exit(2)
	}
	index, err := strconv.Atoi(rest[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid index %q\n", rest[0])
		os.Exit(2)
	}
	ok, err := storage.DeleteExpense(index)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if ok {
		fmt.Printf(" Deleted index %d\n", index)
	} else {
		fmt.Printf(" Invalid index %d (use list --show-index)\n", index)
	}
}
