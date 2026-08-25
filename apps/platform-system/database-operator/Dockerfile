FROM python:3.11-slim

WORKDIR /app

# Upgrade base tooling to clear vendored setuptools/wheel CVEs
RUN python -m pip install --no-cache-dir --upgrade pip setuptools wheel

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Create non-root system group and user (Debian/Ubuntu syntax)
RUN groupadd --system appgroup && useradd --system --gid appgroup appuser

COPY . .

# Switch ownership and switch user
RUN chown -R appuser:appgroup /app

USER appuser

CMD ["python", "main.py"]
