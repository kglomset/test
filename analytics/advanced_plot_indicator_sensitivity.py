#!/usr/bin/env python3
"""Plot indicator feature sensitivities for a given product grouped by theme."""

import argparse
import json

import matplotlib.pyplot as plt
import numpy as np


def load_model(model_file: str):
    """
    Load model parameters, metadata, and product names from JSON.
    Returns indicator features list, beta dict, and names dict.
    """
    with open(model_file) as f:
        data = json.load(f)
    meta = data.pop("metadata")
    indicators = meta["dummy_features"]
    beta = {}
    names = {}
    for pid_str, info in data.items():
        pid = int(pid_str)
        beta[pid] = info["beta"]
        names[pid] = info.get("name", str(pid))
    return indicators, beta, names


def group_indicators(indicators):
    """
    Group indicator features by theme based on prefix.
    Themes: 'clouds', 'snow_type', 'track', others.
    """
    groups = {}
    for feat in indicators:
        if feat.startswith("clouds_"):
            groups.setdefault("clouds", []).append(feat)
        elif feat.startswith("snow_type_"):
            groups.setdefault("snow_type", []).append(feat)
        elif feat.startswith("track_"):
            groups.setdefault("track", []).append(feat)
        else:
            groups.setdefault("other", []).append(feat)
    return groups


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--model_file",
        required=True,
        help="Path to trained product_model.json"
    )
    parser.add_argument(
        "--product",
        type=int,
        required=True,
        help="Product ID to plot sensitivities for"
    )
    parser.add_argument(
        "--output_file",
        default="indicator_sensitivity.png",
        help="Path to save indicator sensitivity plot"
    )
    args = parser.parse_args()

    indicators, beta_dict, names = load_model(args.model_file)
    if args.product not in beta_dict:
        raise ValueError(f"Product ID {args.product} not found in model")

    groups = group_indicators(indicators)
    n_groups = len(groups)

    fig, axes = plt.subplots(1, n_groups, figsize=(5 * n_groups, 4), squeeze=False)

    for idx, (theme, feats) in enumerate(groups.items()):
        # compute absolute beta sensitivities for this product
        vals = [abs(beta_dict[args.product].get(f, 0.0)) for f in feats]
        # sort within group
        if feats:
            sorted_feats, sorted_vals = zip(
                *sorted(zip(feats, vals), key=lambda x: x[1], reverse=True)
            )
        else:
            sorted_feats, sorted_vals = [], []

        y_pos = np.arange(len(sorted_feats))
        ax = axes[0, idx]
        ax.barh(y_pos, sorted_vals)
        ax.set_yticks(y_pos)
        ax.set_yticklabels(sorted_feats)
        ax.invert_yaxis()
        ax.set_xlabel("Sensitivity (|beta|)")
        ax.set_title(theme.replace("_", " ").title())

    fig.suptitle(f"Indicator sensitivities for {names[args.product]}")
    fig.tight_layout(rect=[0, 0, 1, 0.96])
    plt.savefig(args.output_file)
    print(f"Saved indicator sensitivity plot to '{args.output_file}'")


if __name__ == "__main__":
    main()
