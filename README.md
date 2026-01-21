# PET - Personal Expense Tracker

Simple CLI tool to track personal expenses. Stores data in `expenses.csv`.

## Features
- Add expenses (amount, description, category, optional date)
- List expenses (with optional filters: month/year/category)
- View summary (total, average, min/max, category breakdown)
- Delete expenses by index

## Installation
```bash
git clone https://github.com/rwenh/PET.git
cd PET
# optional: python3 -m venv venv && source venv/bin/activate
pip install -r requirements.txt   # (empty for now – no deps)
python3 pet.py add 250 "Lunch" -c Food
python3 pet.py list
python3 pet.py list --show-index
python3 pet.py list --month 1 --year 2026
python3 pet.py summary
python3 pet.py summary --month 12
python3 pet.py delete 3
python3 pet.py --help
