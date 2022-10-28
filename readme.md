# Phi-Accrual Failure Detection

Implementation of a phi-accrual failure detector from `Hayashibara et al.` with toy client and server.

## Where It's Used

* [Akka](https://doc.akka.io/docs/akka/current/typed/failure-detector.html)
* [Cassandra](https://docs.datastax.com/en/cassandra-oss/2.2/cassandra/architecture/archDataDistributeFailDetect.html)

## Running on DO

```bash
cd ./demo/infra && terraform apply -var digitalocean_token=$DIGITALOCEAN_TOKEN
```

## Running Locally

```bash
# build
docker build . -f ./examples/client/Dockerfile -t dmw2151/phi-failure-client
docker build . -f ./examples/server/Dockerfile -t dmw2151/phi-failure-server

# start client && server communicating over localhost...
docker run --rm --name phi-failure-server \
	-p 52150:52150 -p 52151:52151 \
	-e METRICS_SERVE_ADDR="0.0.0.0:52150"\
	-e FAILURE_DETECTOR_LISTEN_ADDR="0.0.0.0:52151" \
	dmw2151/phi-failure-server ./server 

docker run --rm --net host --name phi-failure-client \
	-e FAILURE_DETECTOR_SERVER_HOST="localhost"\
	-e FAILURE_DETECTOR_SERVER_PORT="52151"\
	dmw2151/phi-failure-client ./client

# visit metrics endpoint @ localhost:52150/metrics 
```

Then visit metrics endpoint @ `localhost:52150/metrics`, should show `failure_detector_suspicion_level`, `failure_detector_heartbeat_interval`, and `failure_detector_heartbeat_interval_stdev` for each client.

```bash
# HELP failure_detector_suspicion_level calculated suspicion level
# TYPE failure_detector_suspicion_level gauge
failure_detector_suspicion_level{client_host_id="39e59da2ccd3",client_pid="1",client_process_uuid="2911ae8d-92f3-48ec-a36e-63b9a82ccb50",client_region="nyc1",server_host_id="9cd2bb3de19f"} 1.0640600229526151
```

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