FROM python:3.11-slim

WORKDIR /app

RUN pip install --no-cache-dir kopf kubernetes

COPY operator.py /app/operator.py

CMD ["kopf", "run", "--all-namespaces", "/app/operator.py"]
