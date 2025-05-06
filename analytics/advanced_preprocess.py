#!/usr/bin/env python3
"""Preprocess raw tests/results into pairwise CSV for Bradley-Terry training."""

import argparse

import pandas as pd

from common import process_test_data


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--input_file",
        default="normalized_products_and_product_tests.xlsx",
        help="Excel file with sheets: tests, results, products"
    )
    parser.add_argument(
        "--output_file",
        default="product_pairs.csv",
        help="Output path for pairwise comparisons CSV"
    )
    args = parser.parse_args()

    tests = pd.read_excel(args.input_file, sheet_name="tests")
    results = pd.read_excel(args.input_file, sheet_name="results")
    products = pd.read_excel(args.input_file, sheet_name="products")

    clean_tests = process_test_data(tests)

    merged = (
        results
        .merge(clean_tests, on="test_id", how="inner")
        .merge(products, on="product_id", how="inner")
    )

    pairs = []
    for test_id, grp in merged.groupby("test_id"):
        grp = grp.sort_values("rank")
        for i in range(len(grp) - 1):
            winner = grp.iloc[i]
            for j in range(i + 1, len(grp)):
                loser = grp.iloc[j]
                entry = {
                    "test_id": test_id,
                    "winner_id": winner["product_id"],
                    "winner_name": winner["product_name"],
                    "loser_id": loser["product_id"],
                    "loser_name": loser["product_name"],
                }
                for col in clean_tests.columns:
                    if col != "test_id":
                        entry[col] = winner[col]
                pairs.append(entry)

    pd.DataFrame(pairs).to_csv(args.output_file, index=False)
    print(f"Saved pairwise comparisons to '{args.output_file}'")


if __name__ == "__main__":
    main()
