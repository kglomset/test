#!/usr/bin/env python3
"""Preprocess raw tests/results into parquet for naive prediction."""

import argparse
import os
import sys

import pandas as pd

from common import DATA_FILE, TESTS_PARQUET, RESULTS_PARQUET, PRODUCTS_PARQUET, process_test_data


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--input_file",
        default=DATA_FILE,
        help="Excel file with sheets: tests, results, products"
    )
    parser.add_argument(
        "--output_dir",
        default=".",
        help="Directory to write parquet files"
    )
    args = parser.parse_args()

    if not os.path.exists(args.input_file):
        print(f"Error: Source file not found: {args.input_file}", file=sys.stderr)
        sys.exit(1)

    tests = pd.read_excel(args.input_file, sheet_name="tests")
    results = pd.read_excel(args.input_file, sheet_name="results")
    products = pd.read_excel(args.input_file, sheet_name="products")

    clean_tests = process_test_data(tests)
    if clean_tests.empty:
        print("Error: No valid test records after filtering", file=sys.stderr)
        sys.exit(1)

    if not {"test_id", "product_id", "rank"}.issubset(results.columns):
        print("Error: Results missing required columns", file=sys.stderr)
        sys.exit(1)
    if not {"product_id", "product_name"}.issubset(products.columns):
        print("Error: Products missing required columns", file=sys.stderr)
        sys.exit(1)

    paths = {
        TESTS_PARQUET: os.path.join(args.output_dir, TESTS_PARQUET),
        RESULTS_PARQUET: os.path.join(args.output_dir, RESULTS_PARQUET),
        PRODUCTS_PARQUET: os.path.join(args.output_dir, PRODUCTS_PARQUET),
    }
    clean_tests.to_parquet(paths[TESTS_PARQUET], index=False)
    results.to_parquet(paths[RESULTS_PARQUET], index=False)
    products.to_parquet(paths[PRODUCTS_PARQUET], index=False)

    print(f"Wrote cleaned tests to '{paths[TESTS_PARQUET]}'")
    print(f"Wrote results to '{paths[RESULTS_PARQUET]}'")
    print(f"Wrote products to '{paths[PRODUCTS_PARQUET]}'")


if __name__ == "__main__":
    main()
