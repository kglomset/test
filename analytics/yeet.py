#!/usr/bin/env python3
"""Plot signed utility contributions (β) for one-hot features for a given product, grouped by theme."""

import argparse
import json
import matplotlib.pyplot as plt
import numpy as np

def load_model(model_file):
    with open(model_file) as f:
        data = json.load(f)
    meta = data.pop("metadata")
    dummy_feats = meta["dummy_features"]
    beta = {}
    names = {}
    for pid_str, info in data.items():
        pid = int(pid_str)
        beta[pid] = info["beta"]
        names[pid] = info.get("name", pid_str)
    return dummy_feats, beta, names

def group_indicators(indicators):
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
    parser.add_argument("--model_file", required=True, help="Path to product_model.json")
    parser.add_argument("--product", type=int, required=True, help="Product ID to plot utility for")
    parser.add_argument("--output_file", default="indicator_utility.pdf", help="Path to save the plot")
    args = parser.parse_args()

    indicators, beta_dict, names = load_model(args.model_file)
    if args.product not in beta_dict:
        raise ValueError(f"Product ID {args.product} not found in model")

    groups = group_indicators(indicators)
    n_groups = len(groups)
    fig, axes = plt.subplots(1, n_groups, figsize=(5 * n_groups, 4), squeeze=False)

    for idx, (theme, feats) in enumerate(groups.items()):
        vals = [beta_dict[args.product].get(f, 0.0) for f in feats]
        # sort by descending signed value
        pairs = sorted(zip(feats, vals), key=lambda x: x[1], reverse=True)
        sorted_feats, sorted_vals = zip(*pairs) if pairs else ([], [])
        y_pos = np.arange(len(sorted_feats))
        ax = axes[0, idx]
        ax.barh(y_pos, sorted_vals)
        ax.set_yticks(y_pos)
        ax.set_yticklabels(sorted_feats)
        ax.invert_yaxis()
        ax.set_xlabel("Utility contribution (β)")
        ax.set_title(theme.replace('_', ' ').title())

    fig.suptitle(f"Indicator Utility for {names[args.product]}")
    fig.tight_layout(rect=[0, 0, 1, 0.96])
    plt.savefig(args.output_file)
    print(f"Saved indicator utility plot to '{args.output_file}'")

if __name__ == "__main__":
    main()
