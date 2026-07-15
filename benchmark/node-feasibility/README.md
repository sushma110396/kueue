# Node Capacity Feasibility Benchmark

## Objective

Evaluate the impact of a pre-admission node capacity feasibility check added to the Kueue scheduler.

The feature rejects workloads whose CPU or memory requests exceed the allocatable resources of every node in the cluster. By preventing these workloads from entering the admission pipeline, the scheduler avoids reserving ClusterQueue quota for workloads that can never be scheduled.

---

## Test Environment

| Component | Value |
|----------|-------|
| Kubernetes | Kind |
| Kueue | v0.19.0-devel |
| Nodes | 1 |
| Node allocatable CPU | 10 CPU |
| ClusterQueue nominal quota | 100 CPU |
| ResourceFlavor | default-flavor |

---

## Benchmark Configuration

### Workloads

Three mixed-workload scenarios were evaluated.

| Scenario | Unschedulable Workloads | Runnable Workloads |
|----------|------------------------:|-------------------:|
| A | 10 | 50 |
| B | 30 | 30 |
| C | 50 | 10 |

Each runnable workload requested **2 CPU**, while each unschedulable workload requested **12 CPU**. Since the cluster contained a single node with **10 allocatable CPU**, unschedulable workloads could never be placed on any node.

### Submission Order

For every benchmark:

1. Submit all unschedulable workloads.
2. Submit all runnable workloads.

This simulates a realistic scenario where invalid workloads enter the queue before runnable workloads, competing for the same ClusterQueue quota.

---

## Benchmark Results

The benchmark compares the baseline scheduler with the modified scheduler across all three workload mixes.

The accompanying charts summarize:

- Runnable workload admissions
- ClusterQueue quota reserved by unschedulable workloads
- Runnable workload admission rate

---

## Representative Result (30 Unschedulable / 30 Runnable)

### Baseline Scheduler (Node Feasibility Check Disabled)

| Metric | Value |
|--------|------:|
| Runnable workloads submitted | 30 |
| Runnable workloads admitted | 2 |
| Unschedulable workloads submitted | 30 |
| Unschedulable workloads admitted | 8 |
| ClusterQueue quota | 100 CPU |
| Quota reserved by unschedulable workloads | 96 CPU |
| Quota reserved by runnable workloads | 4 CPU |

### Observation

The baseline scheduler admitted unschedulable workloads until ClusterQueue quota was nearly exhausted.

Eight unschedulable workloads reserved **96 CPU**, leaving only **4 CPU** available. Consequently, only **2 of 30 runnable workloads** could be admitted despite being schedulable.

---

### Modified Scheduler (Node Feasibility Check Enabled)

| Metric | Value |
|--------|------:|
| Runnable workloads submitted | 30 |
| Runnable workloads admitted | 30 |
| Unschedulable workloads submitted | 30 |
| Unschedulable workloads admitted | 0 |
| ClusterQueue quota | 100 CPU |
| Quota reserved by unschedulable workloads | 0 CPU |
| Quota reserved by runnable workloads | 60 CPU |

### Observation

The node feasibility check rejected unschedulable workloads before quota reservation.

As a result:

- All runnable workloads were admitted.
- No ClusterQueue quota was reserved by unschedulable workloads.
- The scheduler avoided creating Pods that could never be scheduled.

---

## Comparison (30 Unschedulable / 30 Runnable)

| Metric | Baseline | Modified |
|--------|---------:|---------:|
| Runnable workloads admitted | 2 | 30 |
| Unschedulable workloads admitted | 8 | 0 |
| ClusterQueue quota reserved by unschedulable workloads | 96 CPU | 0 CPU |
| ClusterQueue quota available for runnable workloads | 4 CPU | 60 CPU |

---

## Key Findings

- Increased runnable workload admission rate from 4–20% in the baseline scheduler to 100% across all benchmark scenarios.
- In the baseline scheduler, unschedulable workloads reserved 96% of the available ClusterQueue quota before runnable workloads could be admitted. The modified scheduler eliminated this wasted reservation, making the full quota available for runnable workloads.
- Prevented futile Pod scheduling attempts for workloads that exceed the allocatable CPU or memory capacity of every node in the cluster, preventing FailedScheduling events by avoiding Pod creation.

---

## Limitations

- Benchmarks were evaluated on a single-node Kind cluster with homogeneous node capacity.
- The feasibility check evaluates CPU and memory requests against node allocatable resources.
- Other scheduling constraints (for example, node affinity, taints, tolerations, topology constraints, and storage availability) were outside the scope of this benchmark.
- Scheduler admission latency and throughput were not measured.
- Future work includes evaluating heterogeneous multi-node clusters and benchmarking scheduler latency under larger workloads.