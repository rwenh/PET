package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// dateLayout is the on-disk date format
const dateLayout = "2006-01-02"

type Expense struct {
	Date        string
	Amount      float64
	Category    string
	Description string
}

type ExpenseStorage struct {
	filepath string
	headers  []string
}

func NewExpenseStorage(path string) (*ExpenseStorage, error) {
	s := &ExpenseStorage{
		filepath: path,
		headers:  []string{"date", "amount", "category", "description"},
	}
	if err := s.ensureFileExists(); err != nil {
		return nil, fmt.Errorf("error initializing %s: %w", path, err)
	}
	return s, nil
}

func (s *ExpenseStorage) ensureFileExists() error {
	if _, err := os.Stat(s.filepath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	f, err := os.Create(s.filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.UseCRLF = true
	defer w.Flush()
	return w.Write(s.headers)
}

func (s *ExpenseStorage) SaveExpense(amount float64, category, description, date string) error {
	if date == "" {
		date = time.Now().Format(dateLayout)
	}

	f, err := os.OpenFile(s.filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error saving expense: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.UseCRLF = true
	defer w.Flush()

	row := []string{date, fmt.Sprintf("%.2f", amount), category, description}
	if err := w.Write(row); err != nil {
		return fmt.Errorf("error saving expense: %w", err)
	}
	return nil
}

func (s *ExpenseStorage) LoadExpenses() ([]Expense, error) {
	f, err := os.Open(s.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Expense{}, nil
		}
		return nil, fmt.Errorf("error loading expenses: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error loading expenses: %w", err)
	}
	if len(records) == 0 {
		return []Expense{}, nil
	}

	expenses := make([]Expense, 0, len(records)-1)
	for _, rec := range records[1:] {
		if len(rec) < 4 || rec[1] == "" {
			continue
		}
		amount, err := strconv.ParseFloat(rec[1], 64)
		if err != nil {
			continue
		}
		expenses = append(expenses, Expense{
			Date:        rec[0],
			Amount:      amount,
			Category:    rec[2],
			Description: rec[3],
		})
	}
	return expenses, nil
}

func (s *ExpenseStorage) DeleteExpense(index int) (bool, error) {
	expenses, err := s.LoadExpenses()
	if err != nil {
		return false, err
	}
	if index < 0 || index >= len(expenses) {
		return false, nil
	}

	expenses = append(expenses[:index], expenses[index+1:]...)
	if err := s.overwriteExpenses(expenses); err != nil {
		return false, err
	}
	return true, nil
}

func (s *ExpenseStorage) overwriteExpenses(expenses []Expense) error {
	f, err := os.Create(s.filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.UseCRLF = true
	defer w.Flush()

	if err := w.Write(s.headers); err != nil {
		return err
	}
	for _, e := range expenses {
		row := []string{e.Date, fmt.Sprintf("%.2f", e.Amount), e.Category, e.Description}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}
