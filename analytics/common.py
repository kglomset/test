#!/usr/bin/env python3
"""Common utilities and constants used by preprocessing and prediction scripts."""

import os

import numpy as np
import pandas as pd

DATA_FILE = "normalized_products_and_product_tests.xlsx"
TESTS_PARQUET = "data/cleaned_tests.parquet"
RESULTS_PARQUET = "data/cleaned_results.parquet"
PRODUCTS_PARQUET = "data/cleaned_products.parquet"

DEFAULT_VALUES = {
    "air_humidity": 72,
    "wind": 1,
    "snow_moisture": 24,
    "clouds_partly_cloudy": 1,
    "clouds_clear_sky": 0,
    "clouds_cloudy": 0,
    "clouds_fog": 0,
    "track_none": 1,
    "track_T1": 0,
    "track_T2": 0,
    "track_D1": 0,
    "track_D2": 0,
}
REQUIRED_PARAMS = ["air_temp", "hardness"]

ORDINAL_RANGES = {
    "air_temp":      (-40, 10),
    "snow_temp":     (-50, 0),
    "hardness":      (1,   6),
    "wind":          (0,   3),
    "snow_moisture": (0, 100),
    "air_humidity":  (0, 100),
}

VALID_SNOW_TYPE = {"A1", "A2", "A3", "A4", "A5", "FS", "NS", "IN", "IT", "TR"}
VALID_HARDNESS = {"H1", "H2", "H3", "H4", "H5", "H6"}
VALID_WIND = {"S", "L", "M", "ST"}
VALID_TRACK = {"none", "T1", "T2", "D1", "D2"}
VALID_CLOUD = {"clear_sky", "partly_cloudy", "cloudy", "fog"}

HARDNESS_MAP = {f"H{i}": i for i in range(1, 7)}
WIND_MAP = {"S": 0, "L": 1, "M": 2, "ST": 3}
SNOW_MOISTURE_MAP = {"DS": 11, "W1": 26, "W2": 43, "W3": 65, "W4": 88}


def normalize_to_unit(x: float, mn: float, mx: float) -> float:
    """Normalize x to [0,1] given range mn to mx."""
    return 0 if mx == mn else round((x - mn) / (mx - mn), 3)


def process_test_data(tests_df: pd.DataFrame) -> pd.DataFrame:
    """
    Clean and normalize raw test data.

    Filters invalid rows, fills defaults, scales numeric features,
    and one-hot encodes categorical columns.
    """
    df = tests_df.copy()
    df = df[
        df["air_temp"].between(*ORDINAL_RANGES["air_temp"]) &
        df["snow_type"].isin(VALID_SNOW_TYPE) &
        df["hardness"].isin(VALID_HARDNESS)
    ]

    df["air_humidity"] = (
        df["air_humidity"]
        .fillna(DEFAULT_VALUES["air_humidity"])
        .clip(0, 100)
    )
    df["clouds"] = df["clouds"].fillna("partly_cloudy")
    df["wind"] = df["wind"].fillna("none")
    df["track"] = df["track"].fillna("none")

    df["snow_moisture"] = df["snow_moisture"].apply(
        lambda x: (
            SNOW_MOISTURE_MAP.get(x, DEFAULT_VALUES["snow_moisture"])
            if isinstance(x, str)
            else (DEFAULT_VALUES["snow_moisture"] if pd.isna(x) else x)
        )
    ).clip(0, 100)

    df["snow_temp"] = df.apply(
        lambda r: r["air_temp"] - 2 if pd.isna(r["snow_temp"]) else r["snow_temp"],
        axis=1,
    ).clip(*ORDINAL_RANGES["snow_temp"])

    df["air_temp"] = df["air_temp"].apply(
        lambda v: normalize_to_unit(v, *ORDINAL_RANGES["air_temp"])
    )
    df["snow_temp"] = df["snow_temp"].apply(
        lambda v: normalize_to_unit(v, *ORDINAL_RANGES["snow_temp"])
    )
    df["hardness"] = (
        df["hardness"].map(HARDNESS_MAP)
        .apply(lambda v: normalize_to_unit(v, *ORDINAL_RANGES["hardness"]))
    )

    df["wind_numeric"] = df["wind"].map(WIND_MAP).fillna(
        DEFAULT_VALUES["wind"]
    )
    df["wind"] = df["wind_numeric"].apply(
        lambda v: normalize_to_unit(v, *ORDINAL_RANGES["wind"])
    )
    df.drop(columns=["wind_numeric"], inplace=True)

    for cat in VALID_CLOUD:
        df[f"clouds_{cat}"] = (df["clouds"] == cat).astype(int)
    for cat in VALID_SNOW_TYPE:
        df[f"snow_type_{cat}"] = (df["snow_type"] == cat).astype(int)
    for cat in VALID_TRACK:
        df[f"track_{cat}"] = (df["track"] == cat).astype(int)

    df.drop(columns=["clouds", "track", "snow_type"], inplace=True)
    return df


__all__ = [
    "DATA_FILE",
    "TESTS_PARQUET",
    "RESULTS_PARQUET",
    "PRODUCTS_PARQUET",
    "DEFAULT_VALUES",
    "REQUIRED_PARAMS",
    "VALID_SNOW_TYPE",
    "normalize_to_unit",
    "process_test_data",
]
