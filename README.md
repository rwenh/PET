# PET - Personal Expense Tracker (Go)

Go port of the Python PET CLI. Stores data in `expenses.csv`, in the same
format the Python version uses — CSV files are interchangeable between the
two.

## Features
- Add expenses (amount, description, category, optional date)
- List expenses (with optional filters: month/year/category)
- View summary (total, average, min/max, category breakdown)
- Delete expenses by index

## Build
```bash
go build -o pet .
```
No external dependencies — standard library only, same as the Python
version's empty `requirements.txt`.

## Usage
```bash
./pet add 420 "Lunch" -c Food
./pet list
./pet list --show-index
./pet list --month 1 --year 2026
./pet summary
./pet summary --month 12
./pet delete 3
./pet --help
./pet add -h
```

## Notes on this port
- **Flags work in any position**, e.g. `pet add 420 "Lunch" -c Food` — Go's
  `flag` package normally requires flags before positional arguments; this
  is worked around with a small pre-parse step so the CLI syntax matches
  the Python version exactly.
- **CSV format is byte-compatible**: `expenses.csv` written by either
  version is interchangeable (same headers, same `\r\n` line endings from
  Python's `csv` module default, same `%.2f` amount formatting).
- **More defensive parsing than the original**: a malformed row (bad
  amount, bad date) is skipped rather than aborting the entire load, and
  `-d/--date` on `add` is validated up front instead of failing silently
  later.
- **Fixed a formatting inconsistency**: in the Python version, deleting an
  expense rewrites the CSV without reapplying `.2f` formatting, so amounts
  can drift (e.g. `420.00` becomes `420.0`). The Go version formats
  consistently on every write.
- Preserved as-is from the original: when `--show-index` is combined with
  `--month`/`--category`, the index shown is the position within the
  *filtered* list, not the full dataset — so it won't line up with
  `delete` unless you're looking at the unfiltered list. This was true in
  the Python version too.
