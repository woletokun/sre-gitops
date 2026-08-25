FROM python:3.11-slim

WORKDIR /app

RUN pip install --no-cache-dir kopf kubernetes

COPY operator.py /app/operator.py

# Create app group and user
RUN addgroup --system appgroup && adduser --system --ingroup appgroup appuser

CMD ["kopf", "run", "--all-namespaces", "/app/operator.py"]
