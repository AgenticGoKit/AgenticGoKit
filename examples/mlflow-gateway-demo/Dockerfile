FROM python:3.10-slim

RUN pip install mlflow[gateway]

WORKDIR /app
COPY gateway-config.yaml /app/

EXPOSE 5001

CMD ["mlflow", "gateway", "start", "--config-path", "/app/gateway-config.yaml", "--port", "5001", "--host", "0.0.0.0"]