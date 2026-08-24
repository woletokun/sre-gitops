import kopf
import kubernetes.client
from kubernetes.client.rest import ApiException

@kopf.on.create('platform.example.com', 'v1alpha1', 'databaseclaims')
def create_db_claim(spec, name, namespace, logger, **kwargs):
    db_name = spec.get('dbName')
    owner_email = spec.get('ownerEmail', 'unassigned')

    logger.info(f"Processing DatabaseClaim '{name}' for DB '{db_name}' (Owner: {owner_email})")

    # Initialize Kubernetes CoreV1Api
    kubernetes.config.load_incluster_config()
    core_api = kubernetes.client.CoreV1Api()

    # Define a Secret containing database credentials for the claimed DB
    secret_manifest = {
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {
            "name": f"{name}-db-creds",
            "namespace": namespace,
            "labels": {
                "app.kubernetes.io/managed-by": "database-claim-operator"
            }
        },
        "stringData": {
            "DATABASE_NAME": db_name,
            "DATABASE_USER": f"user_{db_name}",
            "DATABASE_PASSWORD": "GeneratedPassword123!"
        }
    }

    # Adopt secret so it is automatically deleted when the CRD is deleted
    kopf.adopt(secret_manifest)

    try:
        core_api.create_namespaced_secret(namespace=namespace, body=secret_manifest)
        logger.info(f"Successfully created Secret '{name}-db-creds'")
    except ApiException as e:
        if e.status != 409:  # Ignore if already exists
            raise e

    return {'status': 'Ready', 'database': db_name, 'secretRef': f"{name}-db-creds"}

@kopf.on.delete('platform.example.com', 'v1alpha1', 'databaseclaims')
def delete_db_claim(spec, name, logger, **kwargs):
    logger.info(f"Cleaning up resources for DatabaseClaim '{name}'")
