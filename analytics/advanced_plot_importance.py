#!/usr/bin/env python3
"""Plot top-K numeric feature importance for given product(s) using quadratic BT model."""

import argparse
import json
import sys

import matplotlib.pyplot as plt
import numpy as np


def load_model(model_file: str):
    """
    Load model parameters, metadata, and product names from JSON.
    Returns numeric features list, params dict, and names dict.
    """
    with open(model_file) as f:
        data = json.load(f)
    meta = data.pop("metadata")
    numeric = meta["numeric_features"]
    params = {}
    names = {}
    for pid_str, info in data.items():
        pid = int(pid_str)
        names[pid] = info.get("name", str(pid))
        params[pid] = {
            "m": info["m"],
            "s": info["s"],
            "beta": info["beta"],
            "intercept": info["intercept"]
        }
    return numeric, params, names


def compute_numeric_contributions(pid, numeric, params, baseline=0.5):
    """
    Compute absolute numeric contributions for each feature at given baseline.
    Numeric contribution: s[f] * (baseline - m[f])**2
    """
    p = params[pid]
    contrib = {}
    for feat in numeric:
        diff = baseline - p["m"][feat]
        contrib[feat] = p["s"][feat] * (diff * diff)
    return contrib


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--model_file", required=True,
        help="Path to trained product_model.json"
    )
    parser.add_argument(
        "--product", required=True, nargs='+', type=int,
        help="One or more product IDs to plot importance for"
    )
    parser.add_argument(
        "--top_k", type=int, default=5,
        help="Number of top numeric features to display per product"
    )
    parser.add_argument(
        "--baseline", type=float, default=0.5,
        help="Baseline normalized value at which to compute contributions"
    )
    parser.add_argument(
        "--output_file", default="numeric_importance.png",
        help="Path to save numeric importance plot"
    )
    args = parser.parse_args()

    numeric, params, names = load_model(args.model_file)
    products = args.product

    # Validate requested product IDs
    for pid in products:
        if pid not in params:
            print(f"Error: Product ID {pid} not found in model", file=sys.stderr)
            sys.exit(1)

    # Prepare data for each product
    all_feats = []
    all_vals = []
    for pid in products:
        contrib = compute_numeric_contributions(pid, numeric, params, baseline=args.baseline)
        items = sorted(contrib.items(), key=lambda x: x[1], reverse=True)[:args.top_k]
        feats, vals = zip(*items) if items else ([], [])
        all_feats.append(feats)
        all_vals.append(vals)

    # Plot side-by-side subplots
    n = len(products)
    fig, axes = plt.subplots(1, n, figsize=(4 * n, 0.5 * args.top_k + 1), squeeze=False)
    for i, pid in enumerate(products):
        ax = axes[0, i]
        feats = all_feats[i]
        vals = all_vals[i]
        y_pos = np.arange(len(feats))
        ax.barh(y_pos, vals)
        ax.set_yticks(y_pos)
        ax.set_yticklabels(feats)
        ax.invert_yaxis()
        ax.set_xlabel(f"Contribution at {args.baseline}")
        ax.set_title(f"{names[pid]}")

    plt.tight_layout()
    fig.savefig(args.output_file)
    print(f"Saved numeric importance plot to '{args.output_file}'")


if __name__ == "__main__":
    main()
