#!/usr/bin/env python3
# Train Bradley-Terry model with per-product quadratic utilities.

import argparse
import json

import numpy as np
import pandas as pd
from scipy.optimize import minimize
from scipy.special import expit


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--input_file",
        default="product_pairs.csv",
        help="CSV of pairwise comparisons for training"
    )
    parser.add_argument(
        "--output_file",
        default="product_model.json",
        help="Path to write JSON model"
    )
    parser.add_argument(
        "--reg",
        type=float,
        default=0.1,
        help="L2 regularization weight"
    )
    args = parser.parse_args()

    df = pd.read_csv(args.input_file)

    non_feat = {
        "test_id", "winner_id", "winner_name",
        "loser_id", "loser_name", "date", "place"
    }
    weather_cols = [c for c in df.columns if c not in non_feat]
    dummy_cols = [
        c for c in weather_cols
        if c.startswith(("clouds_", "snow_type_", "track_"))
    ]
    numeric_cols = [c for c in weather_cols if c not in dummy_cols]

    X_num = df[numeric_cols].astype(float).values
    X_dummy = df[dummy_cols].astype(float).values

    product_ids = sorted(set(df["winner_id"]) | set(df["loser_id"]))
    id2idx = {pid: i for i, pid in enumerate(product_ids)}
    w_idx = df["winner_id"].map(id2idx).values
    l_idx = df["loser_id"].map(id2idx).values

    n_prod = len(product_ids)
    n_num = len(numeric_cols)
    n_dummy = len(dummy_cols)
    n_params = n_num * 2 + n_dummy + 1

    def unpack(theta):
        # Slice parameter vector into m, s, beta, intercept.
        theta = theta.reshape(n_prod, n_params)
        m = theta[:, :n_num]
        s = theta[:, n_num : 2 * n_num]
        beta = theta[:, 2 * n_num : 2 * n_num + n_dummy]
        intercept = theta[:, -1]
        return m, s, beta, intercept

    def nll(theta):
        # Negative log-likelihood with L2 penalty.
        m_arr, s_arr, b_arr, iv = unpack(theta)
        dw = (X_num - m_arr[w_idx]) ** 2
        dl = (X_num - m_arr[l_idx]) ** 2

        Uw = -np.sum(s_arr[w_idx] * dw, axis=1)
        Ul = -np.sum(s_arr[l_idx] * dl, axis=1)

        Uw += np.sum(b_arr[w_idx] * X_dummy, axis=1)
        Ul += np.sum(b_arr[l_idx] * X_dummy, axis=1)

        Uw += iv[w_idx]
        Ul += iv[l_idx]

        diff = Uw - Ul
        ll = np.sum(np.log(expit(diff) + 1e-12))
        return -ll + args.reg * np.sum(theta ** 2)

    bounds = []
    for _ in range(n_prod):
        bounds += [(0.0, 1.0)] * n_num
        bounds += [(1e-6, None)] * n_num
        bounds += [(None, None)] * n_dummy
        bounds += [(None, None)]

    rng = np.random.default_rng(17)
    init = np.hstack([
        rng.uniform(0, 1, size=(n_prod, n_num)).ravel(),
        rng.uniform(0.1, 2.0, size=(n_prod, n_num)).ravel(),
        np.zeros((n_prod, n_dummy)).ravel(),
        np.zeros(n_prod)
    ])

    result = minimize(
        nll, init,
        method="L-BFGS-B",
        bounds=bounds,
        options={"maxiter": 1000}
    )
    m_est, s_est, beta_est, iv_est = unpack(result.x)

    model = {}
    for pid, idx in id2idx.items():
        names = (
            df.loc[df["winner_id"] == pid, "winner_name"]
            .dropna()
            .unique()
            .tolist()
        )
        name = names[0] if names else str(pid)
        model[str(pid)] = {
            "id": pid,
            "name": name,
            "m": {f: float(m_est[idx, j]) for j, f in enumerate(numeric_cols)},
            "s": {f: float(s_est[idx, j]) for j, f in enumerate(numeric_cols)},
            "beta": {
                f: float(beta_est[idx, j])
                for j, f in enumerate(dummy_cols)
            },
            "intercept": float(iv_est[idx])
        }

    model["metadata"] = {
        "numeric_features": numeric_cols,
        "dummy_features": dummy_cols
    }

    with open(args.output_file, "w") as f:
        json.dump(model, f, indent=2)

    print(
        f"Trained quadratic BT model: {n_prod} products x "
        f"{n_num} numeric features, {n_dummy} dummies -> "
        f"'{args.output_file}'"
    )


if __name__ == "__main__":
    main()
