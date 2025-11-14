Trying to train a model to classify motion type (stable, shaking, walking, running) based off accelerometer values from LIS3DH

## Development

- **Language / runtime**: Go (`ingestor/go.mod`).
- **MQTT client & broker**: Eclipse Paho client in `internal/mqtt/mqtt.go`; Mosquitto manifests in `k8s/base/mosquitto`.
- **Database**: PostgreSQL with init script `k8s/base/postgres/init.sql` and StatefulSet `k8s/base/postgres/statefulset.yaml`.
- **Metrics / observability**: Prometheus client and HTTP handler in `internal/metrics/metrics.go`.
- **Deployment**: Kubernetes manifests + Kustomize overlays (`k8s/base`, `k8s/overlays/dev/kustomization.yaml`).
- **Build / deps**: Go modules (`ingestor/go.mod`).
- **Configs & secrets**: example secrets in `k8s/base/*/secret.example.yaml` and `ops/secrets/passwd`.

