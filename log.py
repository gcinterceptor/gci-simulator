import logging

def get_logger(path_file, logger_name):
    handler = logging.FileHandler(path_file, mode='w')

    formatter = logging.Formatter('%(asctime)s %(levelname)s %(message)s')
    handler.setFormatter(formatter)

    logger = logging.getLogger(logger_name)
    logger.setLevel(logging.DEBUG)
    logger.addHandler(handler)

    return logger

def log_server_data(logger, server):
    logger.info("Heap level: %.5f%%" % server.heap.level)
    logger.info("Remaining requests in queue: %i" % len(server.queue.items))
    logger.info("GCI executions: %i" % server.gci.times_performed)
    logger.info("GC executions: %i" % server.gc.times_performed)
    logger.info("GC execution time sum: %.3f seconds" % server.gc.gc_exec_time_sum)

def log_latency(logger, requests):
    logger.info("Processed requests: %i" % len(requests))
    logger.info("Latency of the first request: %.3f" % requests[0]._latency_time)
    logger.info("Latency of the last request: %.3f" % requests[-1]._latency_time)
    media = sum(request._latency_time for request in requests) / len(requests)
    logger.info("Average latency of the requests: %.3f" % media)

def log_percentiles(logger, requests, percentile):
    ordered_requests = sorted(requests, key = lambda x: x._latency_time)
    N = len(requests)
    median = percentile(50, N)
    _90th = percentile(90, N)
    _95th = percentile(95, N)
    _99th = percentile(99, N)
    _999th = percentile(99.9, N)
    logger.info("Median: %.5f" % float(ordered_requests[median - 1]._latency_time))
    logger.info("90th Percentile: %.5f" % float(ordered_requests[_90th - 1]._latency_time))
    logger.info("95th Percentile: %.5f" % float(ordered_requests[_95th - 1]._latency_time))
    logger.info("99th Percentile: %.5f" % float(ordered_requests[_99th - 1]._latency_time))
    logger.info("99.9th Percentile: %.5f" % float(ordered_requests[_999th - 1]._latency_time))
