from fastapi import FastAPI
from internal.ml.src.model.predict_model import predict

# Swagger API documentation at /docs#

# Start with command: uvicorn internal.ml.api.main:app --port 8001
app = FastAPI()

@app.get("/")
def root():
    return {"message": "Hello World!"}

# Endpoint to get predictions (For example: http://localhost:8001/predict?rating=4.5)
@app.get("/predict")
def get_prediction(rating: float):
    prediction = predict(rating)
    return {"prediction": prediction}

