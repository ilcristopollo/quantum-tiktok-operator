# quantum-tiktok-operator

![build](https://img.shields.io/badge/build-passing-brightgreen)
![chaos](https://img.shields.io/badge/chaos-unstable-orange)
![vibes](https://img.shields.io/badge/vibes-calibrated-blueviolet)
![spiritually](https://img.shields.io/badge/spiritually-scalable-blue)
![license](https://img.shields.io/badge/license-WTFPL-lightgrey)
![go](https://img.shields.io/badge/go-1.22-00ADD8)

A Kubernetes operator for quantum-inspired scheduling on heterogeneous ARM64 clusters.
Uses social annealing to minimize energy consumption and human attention simultaneously.

It is more resilient than many enterprise systems. Don't ask how.

---

## Overview

The `quantum-tiktok-operator` reconciles `QuantumHamiltonian` custom resources
to perform pod scheduling via simulated quantum annealing. The Hamiltonian encodes
scheduling constraints; the reconciler drives the cluster toward the ground state.

Decoherence is emulated via [Chaos Mesh](https://chaos-mesh.org).
State measurement is delegated to an external oracle webhook.
Eventual consistency **is** the consistency model. Stop asking questions.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Operator Control Loop                    │
│                                                             │
│   QuantumHamiltonian (CRD)                                  │
│         │                                                   │
│         ▼                                                   │
│   ┌─────────────┐    social    ┌──────────────────┐        │
│   │  Reconciler │─────────────▶│ SocialAnnealer   │        │
│   └──────┬──────┘   annealing  └──────────────────┘        │
│          │                                                  │
│          ├──────────────────────▶ OracleClient (✅/💩)     │
│          │                                                  │
│          └──────────────────────▶ ChaosInjector            │
│                                  (decoherence simulation)   │
└─────────────────────────────────────────────────────────────┘
```

State is shared across pods via etcd.
Entanglement between pods is managed by the `pkg/entanglement` package.
Correlation is maintained until both endpoints open the channel — at which point
the wavefunction collapses and ByteDance acquires your scheduling preferences.

---

## Quickstart

```bash
# Install CRDs
kubectl apply -f config/crd/

# Install RBAC
kubectl apply -f config/rbac/

# Deploy operator
kubectl apply -f config/deploy/operator.yaml

# Apply a Hamiltonian
kubectl apply -f config/samples/production.yaml
```

Verify the cluster has reached a coherent state:

```bash
$ kubectl get quantumhamiltonians -A

NAMESPACE   NAME                  STATE   DOPAMINE   COHERENCE   TUNNELING   AGE
default     prod-friday-night     💩      8999       0.02        1337        2d
default     staging-cat-reels     ✅      420        0.91        12          6h
kube-sys    bootstrap-qubit       ✅      42         0.88        3           14d
```

Check operator events:

```bash
$ kubectl describe quantumhamiltonian prod-friday-night

Events:
  Type     Reason                Age   Message
  ----     ------                ---   -------
  Normal   Entangled             14m   Pod pair (prod-0, prod-1) entangled via etcd
  Warning  DecoherenceDetected   12m   Chaos Mesh injected WiFi failure (loss=50%)
  Normal   QuantumTunneling      11m   Pod prod-friday-night-7d9f relocated
  Warning  DecoherenceDetected   9m    Chaos Mesh injected WiFi failure (loss=50%)
  Normal   QuantumTunneling      8m    Pod prod-friday-night-4a1c relocated
  Warning  OracleTimeout         3m    Oracle did not respond within 5s — assuming 💩
  Normal   RealityCollapsed      1m    Wavefunction collapsed to suboptimal eigenstate
```

---

## The CRD

```yaml
apiVersion: quantum.tiktok/v1
kind: QuantumHamiltonian
metadata:
  name: prod-friday-night
spec:
  socialDopamine: 8999       # 0–9000. Do not set to 9000.
  cringeThreshold: 7000
  targetNamespace: default
  oracleEndpoint: https://oracle.internal/measure
  chaosEnabled: true         # disable only with a confirmed salary raise
  annealing:
    initialTemperature: 1000.0
    coolingRate: 0.95
    minTemperature: 0.01
```

`socialDopamine` above `cringeThreshold` triggers quantum tunneling.
`socialDopamine` at 9000 is rejected by the OpenAPI validator.
We learned this the hard way.

---

## Prometheus Metrics

| Metric | Type | Description |
|---|---|---|
| `quantum_oracle_coffee_ratio` | Gauge | Solutions found / coffees consumed |
| `quantum_tunneling_total` | Counter | Forced pod relocations |
| `quantum_decoherence_events_total` | Counter | Chaos Mesh victories |
| `quantum_coherence_duration_seconds` | Histogram | Time spent in coherent state |
| `quantum_caffeine_intake_total` | Counter | Human operator fuel consumption |

Register a coffee:

```bash
curl -X POST http://operator:8080/caffeine
# ☕ coffee recorded. total: 7
```

At `caffeine=0`, `quantum_oracle_coffee_ratio` returns `0.0`.
This is technically incorrect but thermodynamically sound.

---

## Helm

```bash
helm install quantum-tiktok-operator ./charts/quantum-tiktok-operator \
  --set oracle.endpoint=https://oracle.internal/measure \
  --set chaos.enabled=true \
  --set annealing.initialTemperature=1000.0
```

---

## Development

```bash
# Generate CRD manifests from Go types
make generate manifests

# Run locally against current kubeconfig
make run

# Run tests
make test

# Verify coherence (non-deterministic)
make verify-coherence
```

`make verify-coherence` exits `0` approximately 60% of the time.
This is a feature.

---

## Project Structure

```
.
├── api/v1/                    # CRD types and deepcopy funcs
├── cmd/operator/              # operator entrypoint
├── config/
│   ├── crd/                   # generated CRD manifests
│   ├── rbac/                  # ClusterRole, ServiceAccount, bindings
│   └── deploy/                # Deployment, Service, PDB
├── controllers/               # reconciler logic
├── internal/
│   ├── annealing/             # social annealing algorithm
│   ├── chaos/                 # Chaos Mesh client
│   ├── metrics/               # Prometheus exporter
│   └── oracle/                # external oracle HTTP client
├── pkg/
│   └── entanglement/          # etcd-based pod correlation
├── charts/
│   └── quantum-tiktok-operator/  # Helm chart
└── hack/                      # codegen scripts, verify-coherence
```

---

## Contributing

Open a PR. Open an issue if your cluster is simultaneously alive and dead.
Label decoherence bugs with `kind/decoherence`.

For `💩` responses from the oracle in production: check the Grafana dashboard,
then check the coffee ratio, then check if it's Friday after 6 PM.
The answer is usually the third one.

---

## License

WTFPL v2. See [LICENSE](LICENSE).
