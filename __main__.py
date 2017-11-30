from simulator.models import Clients, LoadBalancer, ServerWithGCI, Server
from log import get_logger, log_server_data, log_latency, log_percentiles
from config import get_config
import simpy, os, time, math

def create_directory():
    if not os.path.isdir("logs"):
        os.mkdir("logs")

def percentile(P, N):
    return math.ceil((P / 100) * N)

def main():
    before = time.time()
    create_directory()
    env = simpy.Environment()

    client_conf = get_config('config/clients.ini', 'clients sleep_time-0.00349 create_request_rate-100 max_requests-inf')
    gc_conf = get_config('config/gc.ini', 'gc sleep_time-0.00001 threshold-0.9')
    gci_conf = get_config('config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-2 initial_eget-0.9')
    loadbalancer_conf = get_config('config/loadbalancer.ini', 'loadbalancer sleep_time-0.0035')
    requests_conf = get_config('config/request.ini', 'request service_time-0.0035 memory-0.02')
    server_conf = get_config('config/server.ini', 'server sleep_time-0.00001')

    log_path = 'logs'
    server = ServerWithGCI(env, 1, server_conf, gc_conf, gci_conf, log_path)
    load_balancer = LoadBalancer(env, server, loadbalancer_conf, log_path)
    clients = Clients(env, load_balancer, client_conf, requests_conf, log_path)

    SIM_DURATION_SECONDS = 12
    env.run(until=SIM_DURATION_SECONDS)

    logger = get_logger(log_path + "/main.log", "Main")
    log_server_data(logger, server)
    log_latency(logger, clients.requests)
    log_percentiles(logger, clients.requests, percentile)

    after = time.time()
    logger.info("Time of execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()