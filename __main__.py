from simulator.modules import Clients, LoadBalancer, ServerWithGCI
from utils import getConfig, getLogger
import simpy, os, time

def create_directory():
    directory = "logs"
    if not os.path.isdir(directory):
        os.mkdir(directory)

def log_server_data(logger, server):
    logger.info("Heap level: %.5f%%" % server.heap.level)
    logger.info("Remaining requests in queue: %i" % len(server.queue.items))
    logger.info("GCI executions: %i" % server.gci.times_performed)
    logger.info("GC executions: %i" % server.gc.times_performed)
    logger.info("GC execution time sum: %.3f seconds" % server.gc.gc_exec_time_sum)

def log_latence(logger, requests):
    logger.info("Processed requests: %i" % len(requests))
    logger.info("Latence of the first request: %.3f" % requests[0]._latence_time)
    logger.info("Latence of the last request: %.3f" % requests[-1]._latence_time)
    sumx = sum(request._latence_time for request in requests)
    media = sumx / len(requests)
    logger.info("Latence media of the requests: %.3f" % media)

def main():

    before = time.time()
    create_directory()

    client_conf = getConfig('config/clients.ini', 'clients sleep_time-0.00349 create_request_rate-100 max_requests-inf')
    gc_conf = getConfig('config/gc.ini', 'gc sleep_time-0.00001 threshold-0.9')
    gci_conf = getConfig('config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-2 initial_eget-0.9')
    loadbalancer_conf = getConfig('config/loadbalancer.ini', 'loadbalancer sleep_time-0.0035')
    requests_conf = getConfig('config/request.ini', 'request service_time-0.0035 memory-0.02')
    server_conf = getConfig('config/server.ini', 'server sleep_time-0.00001')

    env = simpy.Environment()

    log_path = 'logs'

    server = ServerWithGCI(env, 1, server_conf, gc_conf, gci_conf, log_path)
    load_balancer = LoadBalancer(env, server, loadbalancer_conf, log_path)

    requests = list()
    Clients(env, load_balancer, requests, client_conf, requests_conf, log_path)

    SIM_DURATION_SECONDS = 12.232
    env.run(until=SIM_DURATION_SECONDS)

    logger = getLogger(log_path + "/main.log", "Main")
    log_server_data(logger, server)
    log_latence(logger, requests)

    after = time.time()
    logger.info("Time of execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()