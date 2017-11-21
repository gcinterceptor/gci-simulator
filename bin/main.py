import sys
sys.path.append("..") # It fixes the problem of imports from modeles

import simpy
from modules import Clients, ServerWithGCI, GCI
from util import getConfig, getLogger

if __name__ == '__main__':

    SIM_DURATION_SECONDS = 12
    env = simpy.Environment()

    client_conf = getConfig('../config/clients.ini', 'clients sleep_time-0.00349 create_request_rate-100 max_requests-inf')
    gc_conf = getConfig('../config/gc.ini', 'gc sleep_time-0.00001 threshold-0.9')
    gci_conf = getConfig('../config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-2 initial_eget-0.0')
    requests_conf = getConfig('../config/request.ini', 'request service_time-0.0035 memory-0.02')
    server_conf = getConfig('../config/server.ini', 'server sleep_time-0.00001')

    requests = list()
    server = ServerWithGCI(env, server_conf, gc_conf, gci_conf)
    clients = Clients(env, server, requests, client_conf, requests_conf)

    env.run(until=SIM_DURATION_SECONDS)

    logger = getLogger("../logs/main.log", "Main")
    logger.info("Heap level: %.5f%%" % server.heap.level)
    logger.info("Remaining requests in queue: %i" % len(server.queue.items))
    logger.info("Processed requests: %i" % len(requests))
    logger.info("GCI executions: %i" % server.gci.times_performed)
    logger.info("GC executions: %i" % server.gc.times_performed)
    logger.info("GC execution time sum: %.3f seconds" % server.gci.gc_exec_time_sum)
    logger.info("Collects performed: %.i" % server.gc.collects_performed)