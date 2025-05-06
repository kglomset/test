#!/usr/bin/env python3
# Predict best product(s) using quadratic BT utility model.

import argparse
import json

import numpy as np

from common import (
    DEFAULT_VALUES,
    REQUIRED_PARAMS,
    HARDNESS_MAP,
    WIND_MAP,
    ORDINAL_RANGES,
    normalize_to_unit,
)


class ProductPredictor:
    # Quadratic utility-based product predictor.

    def __init__(self, model_file: str, importance_scale: float = 1.0):
        with open(model_file) as f:
            data = json.load(f)

        meta = data.pop("metadata")
        self.numeric_features = meta["numeric_features"]
        self.dummy_features = meta["dummy_features"]
        self.importance_scale = importance_scale

        self.product_ids = []
        self.product_names = {}
        self.m = {}
        self.s = {}
        self.beta = {}
        self.intercept = {}

        for pid_str, info in data.items():
            pid = int(pid_str)
            self.product_ids.append(pid)
            self.product_names[pid] = info.get("name", str(pid))
            self.m[pid] = info["m"]
            self.s[pid] = info["s"]
            self.beta[pid] = info["beta"]
            self.intercept[pid] = info["intercept"]

    def prepare_input(self, weather_params: dict) -> dict:
        # Validate and build feature dict from input parameters.
        clean = {k.replace(" *", ""): v for k, v in weather_params.items()}

        # allow string values for hardness and wind, map to normalized floats
        if "hardness" in clean and isinstance(clean["hardness"], str):
            h = HARDNESS_MAP.get(clean["hardness"])
            if h is None:
                raise ValueError(f"Invalid hardness value: {clean['hardness']}")
            clean["hardness"] = normalize_to_unit(h, *ORDINAL_RANGES["hardness"])

        if "wind" in clean and isinstance(clean["wind"], str):
            w = WIND_MAP.get(clean["wind"])
            if w is None:
                raise ValueError(f"Invalid wind value: {clean['wind']}")
            clean["wind"] = normalize_to_unit(w, *ORDINAL_RANGES["wind"])

        for req in REQUIRED_PARAMS:
            if req not in clean:
                raise ValueError(f"Missing required '{req}'")

        if not any(
            clean.get(f, 0) == 1 for f in self.dummy_features
            if f.startswith("snow_type_")
        ):
            raise ValueError("At least one snow_type_ must be set to 1")

        feats = {}
        for feat in self.numeric_features:
            if feat in clean:
                feats[feat] = float(clean[feat])
            elif feat in DEFAULT_VALUES:
                feats[feat] = float(DEFAULT_VALUES[feat])
            elif feat == "snow_temp":
                feats[feat] = clean["air_temp"] - 2
            else:
                feats[feat] = 0.0

        for feat in list(feats):
            mn, mx = ORDINAL_RANGES[feat]
            feats[feat] = normalize_to_unit(feats[feat], mn, mx)

        for feat in self.dummy_features:
            feats[feat] = float(clean.get(feat, DEFAULT_VALUES.get(feat, 0)))

        return feats

    def predict_best_product(
        self,
        weather_params: dict,
        top_n_products: int = None,
        normalize: bool = False
    ) -> dict:
        # Compute and rank products; optionally normalize and limit output.
        feats = self.prepare_input(weather_params)
        results = []

        for pid in self.product_ids:
            util = self.intercept[pid]
            for feat in self.numeric_features:
                diff = feats[feat] - self.m[pid][feat]
                util += -self.s[pid][feat] * diff * diff
            for feat in self.dummy_features:
                util += self.beta[pid][feat] * feats[feat]
            results.append({"id": pid, "name": self.product_names[pid], "score": util})

        results.sort(key=lambda x: x["score"], reverse=True)

        if top_n_products is not None and top_n_products < len(results):
            results = results[:top_n_products]

        if normalize and results:
            scores = [d["score"] for d in results]
            mn, mx = min(scores), max(scores)
            span = mx - mn
            if span > 0:
                for d in results:
                    d["score"] = (d["score"] - mn) / span * 100.0
            else:
                for d in results:
                    d["score"] = 100.0

        for i, d in enumerate(results, start=1):
            d["rank"] = i

        contrib = {}
        if results:
            winner = results[0]["id"]
            for feat in self.numeric_features:
                diff = feats[feat] - self.m[winner][feat]
                contrib[feat] = abs(self.s[winner][feat] * diff * diff)
            for feat in self.dummy_features:
                contrib[feat] = abs(self.beta[winner][feat] * feats[feat])
            for k in contrib:
                contrib[k] *= self.importance_scale
            top_feats = sorted(contrib.items(), key=lambda kv: kv[1], reverse=True)[:5]
        else:
            top_feats = []

        return {
            "ranked_products": results,
            "top_features_for_winner": top_feats
        }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--model_file",
        required=True,
        help="Path to trained product_model.json"
    )
    parser.add_argument(
        "--query_json",
        required=True,
        help="JSON string with environment parameters"
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
        "--importance_scale",
        type=float,
        default=1.0,
        help="Scale feature importances by this factor"
    )
    args = parser.parse_args()

    predictor = ProductPredictor(
        args.model_file, importance_scale=args.importance_scale
    )
    query = json.loads(args.query_json)

    out = predictor.predict_best_product(
        query,
        top_n_products=args.top_n_products,
        normalize=args.normalize
    )

    print("Predicted ranking:")
    for p in out["ranked_products"]:
        norm_part = f", Norm:{p['score']:.2f}" if args.normalize else ""
        print(f"  {p['rank']}. {p['name']}  (ID:{p['id']}, Score:{p['score']:.2f}{norm_part})")

    if args.top_n_products:
        total = len(predictor.product_ids)
        print(f"\nShowing top {args.top_n_products} of {total} products")

    print(f"\nTop features for winner (x{args.importance_scale:.1f}):")
    for feat, val in out["top_features_for_winner"]:
        print(f"  {feat}: {val:.4f}")


if __name__ == "__main__":
    main()
