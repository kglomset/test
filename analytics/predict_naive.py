#!/usr/bin/env python3
# Predict product ranking using weighted aggregation (naive approach) with recency floor.

import argparse
import datetime as dt
import json
import math
import os
import sys

import numpy as np
import pandas as pd

from common import (
    DEFAULT_VALUES,
    HARDNESS_MAP,
    RESULTS_PARQUET,
    REQUIRED_PARAMS,
    TESTS_PARQUET,
    VALID_SNOW_TYPE,
    PRODUCTS_PARQUET,
    WIND_MAP,
    ORDINAL_RANGES,
    normalize_to_unit,
)

WEIGHTS_FILE = "weights_naive.json"
RECENCY_FLOOR = 0.7
TARGET_AT_ONE_YEAR = 0.92  # penalty at age = 365 days
DEFAULT_RECENCY_DAYS = 365


def rank_points(rank: int, n_products: int) -> int:
    # Map rank to score points based on number of products.
    small = {2: [1, 0], 3: [2, 1, 0]}
    if n_products in small:
        tbl = small[n_products]
        return tbl[rank - 1] if 1 <= rank <= n_products else 0
    base = [4, 2, 1, 0, 0, -1, -2]
    return base[rank - 1] if 1 <= rank <= len(base) else -2


def similarity(q_vec, test_vec, w_vec, lambda_days, age_days):
    # Compute weighted distance with exponential recency floor.
    d = math.sqrt(np.sum(w_vec * (q_vec - test_vec) ** 2))
    tau = -lambda_days / math.log((TARGET_AT_ONE_YEAR - RECENCY_FLOOR) / (1.0 - RECENCY_FLOOR))
    decay = math.exp(-age_days / tau)
    penalty = RECENCY_FLOOR + (1.0 - RECENCY_FLOOR) * decay
    return float(d / penalty)


def load_weights(path: str) -> dict:
    # Load feature weights and recency days.
    with open(path) as f:
        w = json.load(f)
    w.setdefault("_recency_days", DEFAULT_RECENCY_DAYS)
    return w


def score_products(
    query: dict,
    tests: pd.DataFrame,
    results: pd.DataFrame,
    products: pd.DataFrame,
    weights: dict,
    closest_tests: int,
    use_all: bool
) -> pd.DataFrame:
    # Score and rank products based on historical tests.
    local_w = weights.copy()
    lambda_days = local_w.pop("_recency_days", DEFAULT_RECENCY_DAYS)

    # Feature columns to consider
    feature_cols = [
        c for c in tests.columns
        if c.startswith((
            "air_", "snow_", "hardness", "wind",
            "clouds_", "track_", "snow_type_"
        ))
    ]

    # Tests are already normalized in [0,1]
    tests_scaled = tests[feature_cols].values

    # Build and normalize the query vector
    q_vals = []
    for c in feature_cols:
        raw = query.get(c, 0)
        if c.startswith(("air_temp","snow_temp","hardness","wind","snow_moisture","air_humidity")):
            key = c if c in ORDINAL_RANGES else c.split("_",1)[1] if "_" in c else c
            mn, mx = ORDINAL_RANGES[key]
            norm = normalize_to_unit(raw, mn, mx)
        else:
            norm = float(raw)
        q_vals.append(norm)
    q_vec = np.array(q_vals)

    # Build weight vector (exclude recency)
    w_vec = np.array([local_w.get(c, 0.0) for c in feature_cols])

    today = dt.date.today()
    distances = []
    for idx, row in tests.iterrows():
        test_date = row["date"].date() if hasattr(row["date"], "date") else row["date"]
        age = (today - test_date).days
        distances.append(similarity(
            q_vec, tests_scaled[idx], w_vec, lambda_days, age
        ))
    tests = tests.assign(sim_distance=distances)

    # Select nearest or all
    selected = tests if use_all else tests.nsmallest(closest_tests, "sim_distance")

    # Aggregate rank-based scores
    scores = {}
    for _, t in selected.iterrows():
        weight = 1 / (1 + t["sim_distance"])
        tid = t["test_id"]
        res = results[results["test_id"] == tid]
        n_prod = len(res)
        for _, r in res.iterrows():
            pts = rank_points(int(r["rank"]), n_prod)
            pid = int(r["product_id"])
            scores[pid] = scores.get(pid, 0.0) + pts * weight

    if not scores:
        return pd.DataFrame(columns=["product_id", "product_name", "score", "score_norm"])

    df_scores = pd.DataFrame(scores.items(), columns=["product_id", "score"])
    mn, mx = df_scores["score"].min(), df_scores["score"].max()
    df_scores["score_norm"] = 100 * (df_scores["score"] - mn) / (mx - mn or 1)
    out = df_scores.merge(products, on="product_id").sort_values("score", ascending=False)

    return out[["product_id", "product_name", "score", "score_norm"]]


def parse_query(raw: str) -> dict:
    # Validate and enrich input query JSON.
    try:
        q = json.loads(raw)
    except json.JSONDecodeError:
        raise ValueError("Invalid JSON for --query_json")

    for req in REQUIRED_PARAMS:
        if req not in q:
            raise ValueError(f"Missing required parameter '{req}'")

    if isinstance(q.get("hardness"), str):
        q["hardness"] = HARDNESS_MAP.get(q["hardness"], int(q["hardness"].lstrip("H")))
    if isinstance(q.get("wind"), str):
        if q["wind"] not in WIND_MAP:
            raise ValueError(f"Invalid wind value: {q['wind']}")
        q["wind"] = WIND_MAP[q["wind"]]

    for col, val in DEFAULT_VALUES.items():
        q.setdefault(col, val)
    q.setdefault("snow_temp", q["air_temp"] - 2)
    snow_flags = [f"snow_type_{t}" for t in VALID_SNOW_TYPE]
    if not any(q.get(f, 0) == 1 for f in snow_flags):
        q["snow_type_TR"] = 1
    return q


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--query_json",
        required=True,
        help="JSON string with current environment"
    )
    parser.add_argument(
        "--closest_tests",
        type=int,
        default=3,
        help="Number of closest tests to include"
    )
    parser.add_argument(
        "--use_all",
        action="store_true",
        help="Aggregate over all historical tests"
    )
    parser.add_argument(
        "--top_n_products",
        type=int,
        default=None,
        help="Number of top products to display"
    )
    parser.add_argument(
        "--normalize",
        action="store_true",
        help="Normalize scores to 0-100"
    )
    parser.add_argument(
        "--weights_file",
        default=WEIGHTS_FILE,
        help="Path to JSON file with feature weights"
    )
    args = parser.parse_args()

    if not os.path.exists(TESTS_PARQUET):
        print("Error: Missing parquet files; run preprocess_naive.py first", file=sys.stderr)
        sys.exit(1)

    tests = pd.read_parquet(TESTS_PARQUET)
    results = pd.read_parquet(RESULTS_PARQUET)
    products = pd.read_parquet(PRODUCTS_PARQUET)

    query = parse_query(args.query_json)
    weights = load_weights(args.weights_file)

    df = score_products(
        query,
        tests,
        results,
        products,
        weights=weights,
        closest_tests=args.closest_tests,
        use_all=args.use_all
    )

    if df.empty:
        print("Error: No matching tests or products found.", file=sys.stderr)
        return

    if args.top_n_products is not None:
        df = df.head(args.top_n_products)

    print("Predicted ranking\n")
    for i, row in enumerate(df.itertuples(), 1):
        norm_part = f", Norm:{row.score_norm:.2f}" if args.normalize else ""
        print(f"  {i}. {row.product_name}  (ID:{row.product_id}, Score:{row.score:.2f}{norm_part})")

    if args.top_n_products:
        total = products["product_id"].nunique()
        print(f"\nShowing top {args.top_n_products} of {total} products")


if __name__ == "__main__":
    main()