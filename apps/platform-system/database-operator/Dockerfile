FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Create non-root system group and user (Debian/Ubuntu syntax)
RUN groupadd -r appgroup && useradd -r -g appgroup appuser

COPY . .

USER appuser

CMD ["python", "main.py"]
