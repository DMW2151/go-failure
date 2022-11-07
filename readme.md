# Phi-Accrual Failure Detection

Implementation of a phi-accrual failure detector from `Hayashibara et al.`.

## Example

Example (see: `./example/` )implements a [lookaside-load-balancer](https://grpc.io/blog/grpc-load-balancing/#lookaside-load-balancing) using a phi-accrual failure detector to serve healthy node addresses.

## Paper

* [The φ accrual failure detector](https://www.researchgate.net/publication/29682135_The_ph_accrual_failure_detector)

```bash
@article{article,
	author = {Hayashibara, Naohiro and Défago, Xavier and Yared, Rami and Katayama, Takuya},
	year = {2004},
	month = {01},
	pages = {},
	title = {The φ accrual failure detector},
	doi = {10.1109/RELDIS.2004.1353004}
}
```
