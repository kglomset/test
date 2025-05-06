#!/usr/bin/env python3
"""Plot quadratic BT utility curves with unified y-scale and optional actual-unit x-axis."""

import argparse
import json

import matplotlib.pyplot as plt
import numpy as np

from common import ORDINAL_RANGES


def load_model(model_file: str):
    """Load model parameters, metadata, and product names from JSON."""
    with open(model_file) as f:
        data = json.load(f)
    meta = data.pop("metadata")
    numeric = meta["numeric_features"]
    dummy = meta["dummy_features"]
    params = {}
    product_names = {}
    for pid_str, info in data.items():
        pid = int(pid_str)
        product_names[pid] = info.get("name", str(pid))
        params[pid] = {
            "m": info["m"],
            "s": info["s"],
            "beta": info["beta"],
            "intercept": info["intercept"],
        }
    return numeric, dummy, params, product_names


def compute_utility(pid, feats, params, numeric, dummy):
    """Compute quadratic utility for a single product."""
    p = params[pid]
    u = p["intercept"]
    for feat in numeric:
        diff = feats[feat] - p["m"][feat]
        u += -p["s"][feat] * diff * diff
    for feat in dummy:
        u += p["beta"][feat] * feats.get(feat, 0.0)
    return u


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--model_file",
        required=True,
        help="Path to product_model.json"
    )
    parser.add_argument(
        "--products",
        nargs="+",
        type=int,
        default=[2, 5, 7, 11],
        help="Product IDs to plot"
    )
    parser.add_argument(
        "--features",
        nargs="+",
        type=int,
        default=[0, 1, 2, 3],
        help="Indices of numeric features to plot"
    )
    parser.add_argument(
        "--n_points",
        type=int,
        default=100,
        help="Number of points per curve"
    )
    parser.add_argument(
        "--range",
        nargs=2,
        type=float,
        default=[0.0, 1.0],
        help="Min and max for numeric features (normalized 0-1)"
    )
    parser.add_argument(
        "--baseline",
        nargs="+",
        type=float,
        help="Baseline values for other numeric features (normalized 0-1, or actual units if --actual-units)"
    )
    parser.add_argument(
        "--actual_units",
        action="store_true",
        help="Display x-axis in actual feature units (uses ORDINAL_RANGES)"
    )
    parser.add_argument(
        "--output_file",
        default="quad_utility_curves.png",
        help="Path to save plot PNG"
    )
    args = parser.parse_args()

    numeric, dummy, params, product_names = load_model(args.model_file)

    # determine baseline normalized values
    if args.baseline:
        if len(args.baseline) != len(numeric):
            raise ValueError("Baseline length must match number of numeric features")
        base_vals = {f: v for f, v in zip(numeric, args.baseline)}
    else:
        mid_norm = 0.5 if args.actual_units else (args.range[0] + args.range[1]) / 2.0
        base_vals = {f: mid_norm for f in numeric}

    # normalized sampling points for utility computation
    xs_norm = (
        np.linspace(0.0, 1.0, args.n_points)
        if args.actual_units
        else np.linspace(args.range[0], args.range[1], args.n_points)
    )

    n_prods = len(args.products)
    n_feats = len(args.features)
    util = np.zeros((n_prods, n_feats, args.n_points))

    # compute utility matrix
    for i, pid in enumerate(args.products):
        for j, fidx in enumerate(args.features):
            vals = []
            feat_name = numeric[fidx]
            for x in xs_norm:
                feats = dict(base_vals)
                feats[feat_name] = x
                for d in dummy:
                    feats.setdefault(d, 0.0)
                vals.append(compute_utility(pid, feats, params, numeric, dummy))
            arr = np.array(vals)
            arr -= arr.max()
            util[i, j, :] = arr

    # helper to get x-values for plotting
    def display_xs(feat_idx):
        feat = numeric[feat_idx]
        if args.actual_units and feat in ORDINAL_RANGES:
            mn, mx = ORDINAL_RANGES[feat]
            return np.linspace(mn, mx, args.n_points)
        return xs_norm

    # plot settings
    #y_min, y_max = util.min(), util.max()
    #pad = (y_max - y_min) * 0.5

    y_min, y_max = util.min(), util.max()
    dr = y_max - y_min
    

    fig, axes = plt.subplots(n_prods, n_feats, figsize=(4 * n_feats, 3 * n_prods), squeeze=False)
    for i, pid in enumerate(args.products):
        for j, fidx in enumerate(args.features):
            ax = axes[i][j]
            xs_plot = display_xs(fidx)
            ax.plot(xs_plot, util[i, j, :], linewidth=2)
            ax.axhline(0, linestyle="--", linewidth=1)
            feat = numeric[fidx]
            # unit label inference
            if "temp" in feat:
                unit = "C"
            elif "humidity" in feat:
                unit = "%"
            else:
                unit = ""
            name = product_names.get(pid, str(pid))
            ax.set_title(f"{name}: {feat}")
            ax.set_xlabel(f"{feat}{(' ' + unit) if unit else ''}")
            ax.set_ylabel("Utility (normalized)")
            #ax.set_ylim(y_min - pad, y_max + pad)
            ax.set_ylim(y_min - 0.1*dr, y_max + 0.8*dr)

    fig.tight_layout()
    fig.savefig(args.output_file)

    print(f"Saved quad utility curves to '{args.output_file}'")


if __name__ == "__main__":
    main()
