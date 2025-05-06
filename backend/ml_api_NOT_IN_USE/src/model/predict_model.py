import numpy as np
import matplotlib.pyplot as plt
import pandas as pd
from sklearn.preprocessing import StandardScaler
from sklearn.svm import SVR

# Code based on Samet Girgin's code on medium.com: 
# https://medium.com/@sametgirgin/support-vector-regression-in-6-steps-with-python-c4569acd062d

# Here i have loded another dataset from the author's GitHub repository: 
# https://github.com/sametgirgin/datasets/blob/main/reviews.csv
dataset = pd.read_csv('internal/ml/src/dataset/reviews.csv')

# Drop missing values
dataset_cleaned = dataset.dropna()

# Extract features (rating) and target variable (reviews)
X = dataset_cleaned.iloc[:, 1:2].values.astype(float)  # Rating as feature
y = dataset_cleaned.iloc[:, 2:3].values.astype(float)  # Reviews as target

# Feature Scaling
sc_X = StandardScaler()
sc_y = StandardScaler()

X_scaled = sc_X.fit_transform(X)
y_scaled = sc_y.fit_transform(y).ravel()  # Flatten y for SVR

# Train SVR model
regressor = SVR(kernel='rbf')
regressor.fit(X_scaled, y_scaled)

# Visualization
#plt.scatter(X_scaled, y_scaled, color='magenta', label="Actual data")
#plt.plot(X_scaled, regressor.predict(X_scaled), color='green', label="SVR model")
#plt.title('Support Vector Regression (Ratings vs Reviews)')
#plt.xlabel('Rating (scaled)')
#plt.ylabel('Reviews (scaled)')
#plt.legend()
#plt.show()

# Output the prediction
def predict(featuredData: float):
    rating_test = np.array([featuredData]).reshape(-1, 1)
    rating_test_scaled = sc_X.transform(rating_test)
    y_pred_scaled = regressor.predict(rating_test_scaled)
    y_pred = sc_y.inverse_transform(y_pred_scaled.reshape(-1, 1))
    return y_pred[0][0]