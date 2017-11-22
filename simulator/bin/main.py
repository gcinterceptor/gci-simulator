from simulator.modules import Clients, ServerWithGCI
from simulator.util import getConfig, getLogger
import simpy

if __name__ == '__main__':

    SIM_DURATION_SECONDS = 12.263
    env = simpy.Environment()

    client_conf = getConfig('../../data/config/clients.ini', 'clients sleep_time-0.00349 create_request_rate-100 max_requests-inf')
    gc_conf = getConfig('../../data/config/gc.ini', 'gc sleep_time-0.00001 threshold-0.9')
    gci_conf = getConfig('../../data/config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-2 initial_eget-0.0')
    requests_conf = getConfig('../../data/config/request.ini', 'request service_time-0.0035 memory-0.02')
    server_conf = getConfig('../../data/config/server.ini', 'server sleep_time-0.00001')

    requests = list()
    server = ServerWithGCI(env, server_conf, gc_conf, gci_conf)
    clients = Clients(env, server, requests, client_conf, requests_conf)

    env.run(until=SIM_DURATION_SECONDS)

    logger = getLogger("../../data/logs/main.log", "Main")
    logger.info("Heap level: %.5f%%" % server.heap.level)
    logger.info("Remaining requests in queue: %i" % len(server.queue.items))
    logger.info("Processed requests: %i" % len(requests))
    logger.info("GCI executions: %i" % server.gci.times_performed)
    logger.info("GC executions: %i" % server.gc.times_performed)
    logger.info("GC execution time sum: %.3f seconds" % server.gc.gc_exec_time_sum)
    logger.info("Collects performed: %.i" % server.gc.collects_performed)