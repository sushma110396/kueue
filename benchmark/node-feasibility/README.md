# Node Capacity-Aware Admission Benchmark

## Objective

Evaluate the impact of introducing node capacity-aware admission into Kueue's scheduler. Before reserving ClusterQueue quota, the scheduler determines whether a workload can fit on at least one node based on allocatable CPU and memory, preventing infeasible workloads from consuming admission resources.

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

Three workload mixes were evaluated to measure how infeasible workloads affect admission decisions under different queue compositions.

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

This models a realistic burst in which infeasible workloads arrive ahead of runnable workloads and compete for the same ClusterQueue quota.

---

## Benchmark Results

The following sections compare the baseline scheduler with the modified scheduler using the representative 30/30 workload mix. Results from all workload mixes are summarized in the charts below.

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

The baseline scheduler admitted infeasible workloads because admission considered only the ClusterQueue quota. Eight workloads reserved 96 CPU, leaving enough quota for only two runnable workloads, even though the remaining runnable workloads could fit on the node.

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

The enhanced admission pipeline identified infeasible workloads before reserving the ClusterQueue quota. As a result, quota remained available for runnable workloads, all runnable workloads were admitted, and unnecessary Pod creation was avoided.

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

- Increased runnable workload admission rates from 4–20% with the baseline scheduler to 100% across all benchmark scenarios.
- Eliminated ClusterQueue quota consumption by infeasible workloads, ensuring quota remained available for runnable workloads.
- Prevented creation of Pods for workloads that exceeded the allocatable CPU or memory capacity of every node, reducing unnecessary scheduling attempts and FailedScheduling events.

---

<img width="2790" height="947" alt="benchmark-charts" src="https://github.com/user-attachments/assets/2bb1ed0b-ca61-4849-8a8d-28adab6545ea" />


## Limitations

- Benchmarks were evaluated on a single-node Kind cluster with homogeneous node capacity.
- The admission logic considered only CPU and memory availability when determining node feasibility.
- Other scheduling constraints (for example, node affinity, taints, tolerations, topology constraints, and storage availability) were outside the scope of this benchmark.
- Scheduler admission latency and throughput were not measured.
- Future work includes evaluating heterogeneous multi-node clusters, incorporating additional scheduling constraints into the feasibility evaluation, and measuring scheduler admission latency under higher workload volumes.
