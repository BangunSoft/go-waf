# GoWaf Service Load Testing Report

## Service Overview
- **Base Image:** Alpine 3.20
- **Go Version:** v1.23
- **CPU Allocation:** 1 vCPU
- **Memory Allocation:** 1 GB
- **Redis Configuration:** 1 vCPU, 1 GB Memory (if using Redis driver)

## Load Testing Configuration
- **Testing Tool:** k6
- **Target File:** Font file in WOFF2 format
- **File Size:** 113.35 KB

### k6 Load Testing Configuration
#### Test Stages
| **Stage** | **Duration** | **Target VUs** |
|-----------|--------------|-----------------|
| Stage 1   | 1 minute     | 1000            |
| Stage 2   | 30 seconds   | 3000            |
| Stage 3   | 1 minute     | 0               |

#### Request Parameters
| **Header**            | **Value**          |
|----------------------|--------------------|
| Connection           | keep-alive         |
| Accept-Encoding      | gzip               |

## Load Test Results

| **Metric**                     | **Redis Driver**                | **Memory Driver**               | **File Driver**                 | **Best Result**                |
|--------------------------------|---------------------------------|---------------------------------|---------------------------------|---------------------------------|
| Execution Duration              | 2m 30s                          | 2m 30s                          | 2m 30s                          | All Equal                      |
| Total VUs                       | 3000                            | 3000                            | 3000                            | All Equal                      |
| Total Iterations Completed       | 41,906                          | 44,352                          | 42,612                          | Memory Driver                  |
| Interrupted Iterations          | 0                               | 0                               | 0                               | All Equal                      |
| Data Received                   | 4.9 GB                          | 5.1 GB                          | 4.9 GB                          | Memory Driver                  |
| Data Sent                       | 7.0 MB                          | 7.4 MB                          | 7.1 MB                          | Memory Driver                  |
| Average Throughput              | 32 MB/s                         | 34 MB/s                         | 33 MB/s                         | Memory Driver                  |
| Total HTTP Requests             | 41,906                          | 44,352                          | 42,612                          | Memory Driver                  |
| Average Request Duration        | 20.6 ms                         | 18.39 ms                       | 19.52 ms                       | Memory Driver                  |
| Minimum Request Duration        | 640 µs                          | 639 µs                          | 647 µs                          | Memory Driver                  |
| Median Request Duration         | 3.27 ms                         | 3.1 ms                          | 3.19 ms                         | Memory Driver                  |
| Maximum Request Duration        | 9.91 ms                         | 9.7 ms                          | 11.44 ms                       | Memory Driver                  |
| 90th Percentile                | 6.12 ms                         | 5.76 ms                         | 5.9 ms                          | Memory Driver                  |
| 95th Percentile                | 167.7 ms                        | 163.11 µs                       | 166.72 µs                       | Memory Driver                  |
| Average Iteration Duration      | 277.51 ms                       | 294.11 ms                       | 282.57 ms                       | Redis Driver                   |
| Maximum Iteration Duration      | 2993 ms                         | 3000 ms                         | 2994 ms                         | Redis Driver                   |

---

## Conclusion
The load tests conducted using the Redis, Memory, and File drivers demonstrate that the service can efficiently handle high traffic loads with 3000 virtual users. The Memory driver consistently outperformed the others in most metrics, while the Redis driver showed the best average and maximum iteration durations.