#!/bin/bash
set -e

# 1. Naive pipeline: parquet preprocessing
python preprocess_naive.py \
  --input_file normalized_products_and_product_tests.xlsx \
  --output_dir .

# 2a. Naive prediction (k=3, all tests, normalized scores)
python predict_naive.py \
  --query_json '{"air_temp":-10,"hardness":"H3","wind":"M","snow_type_TR":1}' \
  --closest_tests 3 \
  --use_all \
  --top_n_products 5 \
  --normalize

# 2b. Naive prediction (k=3, only closest tests, raw scores)
python predict_naive.py \
  --query_json '{"air_temp":-10,"hardness":"H3","wind":"M","snow_type_TR":1}' \
  --closest_tests 3 \
  --top_n_products 5

# 3. Advanced pipeline: CSV pairwise preprocessing
python advanced_preprocess.py \
  --input_file normalized_products_and_product_tests.xlsx \
  --output_file product_pairs.csv

# 4. Advanced training: fit quadratic BT model
python advanced_training_quad.py \
  --input_file product_pairs.csv \
  --output_file product_model.json \
  --reg 0.1

# 5. Inspect first few lines of the model
head -n 20 product_model.json

# 6a. Advanced prediction (normalized scores)
python advanced_predict_quad.py \
  --model_file product_model.json \
  --query_json '{"air_temp":-10,"hardness":"H3","wind":"M","snow_type_TR":1}' \
  --top_n_products 5 \
  --normalize \
  --importance_scale 1.0

# 6b. Advanced prediction (raw scores)
python advanced_predict_quad.py \
  --model_file product_model.json \
  --query_json '{"air_temp":-10,"hardness":"H3","wind":"M","snow_type_TR":1}' \
  --top_n_products 5 \
  --importance_scale 1.0

# 7a. Plot utility curves (normalized baseline at midpoint)
python advanced_plot_quad.py \
  --model_file product_model.json \
  --products 2 5 7 \
  --features 0 1 2 \
  --baseline 0.5 0.5 0.5 0.5 0.5 0.5 \
  --output_file quad_utility_curves.png

# 7b. Plot utility curves (actual units)
python advanced_plot_quad.py \
  --model_file product_model.json \
  --products 2 5 7 \
  --features 0 1 2 \
  --actual_units \
  --baseline -40 -50 3 1 50 50 \
  --output_file quad_utility_curves_actual_units.png

# 8. Plot numeric feature importance (explicit baseline)
python advanced_plot_importance.py \
  --model_file product_model.json \
  --product 2 5 \
  --top_k 5 \
  --baseline 0.5 \
  --output_file numeric_importance.png

# 9. Plot indicator sensitivity
python advanced_plot_indicator_sensitivity.py \
  --model_file product_model.json \
  --product 2 \
  --output_file indicator_sensitivity.png

# 10. Verify outputs
ls quad_utility_curves.png \
   quad_utility_curves_actual_units.png \
   numeric_importance.png \
   indicator_sensitivity.png \
   product_pairs.csv \
   product_model.json

echo "All steps completed successfully."
