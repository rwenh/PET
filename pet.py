#!/usr/bin/env python3
"""
Personal Expense Tracker - CLI Application
Command-line tool to track expenses with CSV storage
"""
from __future__ import annotations

import argparse
import csv
from collections import defaultdict
from datetime import datetime
from pathlib import Path


class ExpenseStorage:
    """Handles CSV read/write operations"""

    def __init__(self, filepath: str = "expenses.csv"):
        self.filepath = Path(filepath)
        self.headers = ["date", "amount", "category", "description"]
        self._ensure_file_exists()

    def _ensure_file_exists(self) -> None:
        if not self.filepath.exists():
            with self.filepath.open("w", newline="", encoding="utf-8") as f:
                csv.DictWriter(f, fieldnames=self.headers).writeheader()

    def save_expense(
        self,
        amount: float,
        category: str,
        description: str,
        date: str | None = None,
    ) -> bool:
        try:
            date = date or datetime.now().strftime("%Y-%m-%d")
            row = {
                "date": date,
                "amount": f"{amount:.2f}",
                "category": category,
                "description": description,
            }
            with self.filepath.open("a", newline="", encoding="utf-8") as f:
                csv.DictWriter(f, fieldnames=self.headers).writerow(row)
            return True
        except Exception as e:
            print(f"Error saving expense: {e}")
            return False

    def load_expenses(self) -> list[dict]:
        try:
            with self.filepath.open(encoding="utf-8") as f:
                reader = csv.DictReader(f)
                return [
                    {**row, "amount": float(row["amount"])}
                    for row in reader
                    if row.get("amount")
                ]
        except FileNotFoundError:
            return []
        except Exception as e:
            print(f"Error loading expenses: {e}")
            return []

    def delete_expense(self, index: int) -> bool:
        expenses = self.load_expenses()
        if 0 <= index < len(expenses):
            expenses.pop(index)
            self._overwrite_expenses(expenses)
            return True
        return False

    def _overwrite_expenses(self, expenses: list[dict]) -> None:
        with self.filepath.open("w", newline="", encoding="utf-8") as f:
            writer = csv.DictWriter(f, fieldnames=self.headers)
            writer.writeheader()
            writer.writerows(expenses)


class ExpenseAnalyzer:
    """Basic expense calculations"""

    @staticmethod
    def filter_by_month(
        expenses: list[dict], month: int, year: int | None = None
    ) -> list[dict]:
        if year is None:
            year = datetime.now().year
        return [
            e
            for e in expenses
            if (d := datetime.strptime(e["date"], "%Y-%m-%d"))
            and d.month == month
            and d.year == year
        ]

    @staticmethod
    def filter_by_category(expenses: list[dict], category: str) -> list[dict]:
        cat = category.lower()
        return [e for e in expenses if e["category"].lower() == cat]

    @staticmethod
    def group_by_category(expenses: list[dict]) -> dict[str, float]:
        groups = defaultdict(float)
        for e in expenses:
            groups[e["category"]] += e["amount"]
        return dict(groups)

    @staticmethod
    def calculate_summary(expenses: list[dict]) -> dict[str, float]:
        if not expenses:
            return {"total": 0, "count": 0, "average": 0, "max": 0, "min": 0}
        amounts = [e["amount"] for e in expenses]
        return {
            "total": sum(amounts),
            "count": len(amounts),
            "average": sum(amounts) / len(amounts),
            "max": max(amounts),
            "min": min(amounts),
        }


class ExpenseFormatter:
    """Output formatting"""

    CURRENCY = "₹"

    @staticmethod
    def format_table(expenses: list[dict], show_index: bool = False) -> None:
        if not expenses:
            print("No expenses found.")
            return

        print("\n" + "-" * 80)
        hdr = f"{'#':<4} " if show_index else ""
        hdr += f"{'Date':<12} {'Amount':>12} {'Category':<15} Description"
        print(hdr)
        print("-" * 80)

        for i, e in enumerate(expenses):
            line = f"{i:<4} " if show_index else ""
            line += (
                f"{e['date']:<12} "
                f"{ExpenseFormatter.CURRENCY}{e['amount']:>9.2f} "
                f"{e['category']:<15} {e['description']}"
            )
            print(line)
        print("-" * 80 + "\n")

    @staticmethod
    def format_summary(summary: dict[str, float], expenses: list[dict]) -> None:
        if not summary["count"]:
            print("No expenses to summarize.")
            return

        c = ExpenseFormatter.CURRENCY
        print("\n" + "=" * 60)
        print("EXPENSE SUMMARY".center(60))
        print("=" * 60)
        print(f"Total:          {c}{summary['total']:>10,.2f}")
        print(f"Count:         {summary['count']:>11}")
        print(f"Average:       {c}{summary['average']:>10,.2f}")
        print(f"Highest:       {c}{summary['max']:>10,.2f}")
        print(f"Lowest:        {c}{summary['min']:>10,.2f}")

        if expenses:
            dates = [datetime.strptime(e["date"], "%Y-%m-%d") for e in expenses]
            days = (max(dates) - min(dates)).days + 1
            daily = summary["total"] / days if days > 0 else 0
            print(f"Daily average: {c}{daily:>10,.2f}")

        print("=" * 60)

        by_cat = ExpenseAnalyzer.group_by_category(expenses)
        if by_cat and summary["total"] > 0:
            print("\nCategory breakdown:")
            print("-" * 60)
            for cat, amt in sorted(by_cat.items(), key=lambda x: x[1], reverse=True):
                pct = amt / summary["total"] * 100
                print(f"{cat:<20} {c}{amt:>11,.2f}  ({pct:5.1f}%)")
            print("-" * 60)


def validate_amount(s: str) -> float:
    try:
        v = float(s)
        if v <= 0:
            raise ValueError
        return v
    except ValueError:
        raise argparse.ArgumentTypeError(f"Invalid positive amount: {s}")


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Personal Expense Tracker",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""examples:
  pet add 420 "Street food" -c Food
  pet list --month 1
  pet summary --month 12 --year 2025
  pet delete 3""",
    )

    sub = parser.add_subparsers(dest="command", required=True)

    add = sub.add_parser("add")
    add.add_argument("amount", type=validate_amount)
    add.add_argument("description")
    add.add_argument("-c", "--category", default="General")
    add.add_argument("-d", "--date")

    lst = sub.add_parser("list")
    lst.add_argument("--month", type=int)
    lst.add_argument("--year", type=int)
    lst.add_argument("-c", "--category")
    lst.add_argument("--show-index", action="store_true")

    smy = sub.add_parser("summary")
    smy.add_argument("--month", type=int)
    smy.add_argument("--year", type=int)
    smy.add_argument("-c", "--category")

    dele = sub.add_parser("delete")
    dele.add_argument("index", type=int)

    args = parser.parse_args()

    storage = ExpenseStorage()
    analyzer = ExpenseAnalyzer()
    formatter = ExpenseFormatter()

    match args.command:
        case "add":
            ok = storage.save_expense(
                args.amount, args.category, args.description, args.date
            )
            print("✓ Added" if ok else "✗ Failed")

        case "list":
            ex = storage.load_expenses()
            if args.month:
                ex = analyzer.filter_by_month(ex, args.month, args.year)
            if args.category:
                ex = analyzer.filter_by_category(ex, args.category)
            formatter.format_table(ex, args.show_index)
            if ex:
                total = sum(e["amount"] for e in ex)
                print(f"Total: {formatter.CURRENCY}{total:,.2f} ({len(ex)} items)")

        case "summary":
            ex = storage.load_expenses()
            if args.month:
                ex = analyzer.filter_by_month(ex, args.month, args.year)
            if args.category:
                ex = analyzer.filter_by_category(ex, args.category)
            sm = analyzer.calculate_summary(ex)
            formatter.format_summary(sm, ex)

        case "delete":
            ok = storage.delete_expense(args.index)
            print(
                f"✓ Deleted index {args.index}"
                if ok
                else f"✗ Invalid index {args.index} (use list --show-index)"
            )


if __name__ == "__main__":
    main()
