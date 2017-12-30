from log import get_logger


class GCSTW(object):

    def __init__(self, env, server, conf, log_path=None):
        self.env = env
        self.server = server
        self.heap = server.heap

        self.threshold = float(conf['threshold'])
        self.sleep_time = float(conf['sleep_time'])
        self.TOTAL_OF_MEMORY = float(conf['TOTAL_OF_MEMORY'])

        self.is_gcing = False
        self.times_performed = 0
        self.collects_performed = 0
        self.gc_exec_time_sum = 0
        self.gc_process = self.env.process(self.run())

        if log_path:
            self.logger = get_logger(log_path + "/gc.log", "GC")
        else:
            self.logger = None

    def run(self):
        while True:
            if self.heap.level >= self.threshold and not self.is_gcing:
                yield self.env.process(self.collect())
                self.times_performed += 1

            yield self.env.timeout(self.sleep_time)  # wait for...

    def collect(self):
        if self.logger:
            self.logger.info(" At %.3f, GC is running. We have %.3f of trash" % (self.env.now, self.heap.level))

        self.is_gcing = True
        self.server.action.interrupt()

        gc_start_time = self.env.now
        while self.heap.level > 0:                                          # while threshold is empty...
            trash = self.heap.level                                         # keeps the amount of trash
            yield self.env.timeout(self.time_by_trash(trash))  # run the time of discarting
            yield self.heap.get(trash)                                      # discards the trash
        self.gc_exec_time_sum += (self.env.now - gc_start_time)

        self.is_gcing = False

        if self.logger:
            self.logger.info(" At %.3f, GC finish his job. Now we have %.3f of trash\n" % (self.env.now, self.heap.level))

        self.server.action = self.env.process(self.server.run())
        self.collects_performed += 1

    def time_by_trash(self, trash):
        return ((trash * self.TOTAL_OF_MEMORY) * 7.317 * (10**-8) + 78.34) / 1000


class GCC(object):

    def __init__(self, env, server, conf, log_path=None):
        self.env = env
        self.server = server
        self.heap = server.heap

        self.threshold = float(conf['threshold'])
        self.sleep_time = float(conf['sleep_time'])
        self.collect_duration = float(conf['collect_duration'])
        self.delay = float(conf['delay'])

        self.is_gcing = False
        self.times_performed = 0
        self.collects_performed = 0
        self.gc_exec_time_sum = 0
        self.gc_process = self.env.process(self.run())

        if log_path:
            self.logger = get_logger(log_path + "/gc.log", "GC")
        else:
            self.logger = None

    def run(self):
        while True:
            if self.heap.level >= self.threshold and not self.is_gcing:
                yield self.env.process(self.collect())
                self.times_performed += 1

            yield self.env.timeout(self.sleep_time)  # wait for...

    def collect(self):
        if self.logger:
            self.logger.info(" At %.3f, GC is running. We have %.3f of trash" % (self.env.now, self.heap.level))

        self.is_gcing, before = True, self.env.now
        trash = self.heap.level
        yield self.heap.get(trash)
        yield self.env.process(self.wait_collect_duration())

        trash = self.heap.level
        if trash > 0:
            yield self.heap.get(trash)
        self.is_gcing, after = False, self.env.now

        if self.logger:
            self.logger.info(" At %.3f, GC finish his job. Now we have %.3f of trash\n" % (self.env.now, self.heap.level))

        self.gc_exec_time_sum += after - before
        self.collects_performed += 1

    def wait_collect_duration(self):
        yield self.env.timeout(self.collect_duration)

    def delay_caused(self):
        return self.delay


class GCI(object):

    def __init__(self, env, server,  conf, log_path=None):
        self.env = env
        self.server = server
        self.check_heap = int(conf['check_heap'])
        self.threshold = float(conf['threshold'])
        self.sleep_time = float(conf['sleep_time'])
        self.estimated_gc_exec_time = float(conf['initial_estimated_gc_exec_time'])

        self.is_shedding = False
        self.HISTORY_SIZE = 5
        self.gcPast = list()
        self.reqPast = list()
        self.processed_requests_history = list()

        self.reqEstimation = 0
        self.requests_to_process = 0
        self.processed_requests = 0
        self.last_processed_requests = 0
        self.times_performed = 0

        if log_path:
            self.logger = get_logger(log_path + "/gci.log", "GCI")
        else:
            self.logger = None

        self._time_shedding = 0

    def before(self, request):
        yield self.env.process(self.check())

        if self.is_shedding:
            if self.logger:
                self.logger.info(" At %.3f, shedding request - shedded request id: %i" % (self.env.now, request.id))
            yield self.env.process(request.load_balancer.shed_request(request, self.server, self._time_shedding))

        else:
            request.arrived_at(self.env.now)
            self.requests_to_process += 1
            yield self.env.process(self.server.enqueue_request(request))

    def check(self):
        if self.processed_requests >= self.check_heap:
            if self.server.heap.level >= self.threshold and not self.is_shedding:
                    self.is_shedding = True
                    self.env.process(self.run_gc())

        yield self.env.timeout(self.sleep_time)

    def run_gc(self):
        if self.logger:
            self.logger.info(" At %.3f, GCI check the heap and it is at %.3f" % (self.env.now, self.server.heap.level))

        self._time_shedding = self.estimated_shed_time()
        # wait for the queue to get empty.
        yield self.env.process(self.check_request_queue())

        # run GC
        gc_start_time = self.env.now
        if self.server.gc.is_gcing:
            while self.server.gc.is_gcing:
                yield self.env.timeout(self.sleep_time)

        elif self.server.heap.level >= self.threshold: # ensure that will only collect if still there is a reason for..
            yield self.env.process(self.server.gc.collect())

        gc_end_time = self.env.now

        # leave server
        self.is_shedding = False
        self._time_shedding = 0

        gc_exec_time = gc_end_time - gc_start_time
        self.update_gci_values(gc_exec_time)

        self.times_performed += 1
        self.requests_to_process = 0
        self.processed_requests = 0

        if self.logger:
            self.logger.info(" At %.3f, GCI finish his job and GC takes %.3f seconds to execute\n" % (self.env.now, gc_exec_time))

    def check_request_queue(self):
        while self.requests_to_process > self.processed_requests:
            yield self.env.timeout(self.sleep_time)
        yield self.env.timeout(self.sleep_time)

    def estimated_shed_time(self):
        return len(self.server.queue.items) * self.reqEstimation + self.estimated_gc_exec_time

    def request_finished(self, request_exe_time):
        self.processed_requests += 1
        if len(self.reqPast) == self.HISTORY_SIZE:
            del self.reqPast[0]
        self.reqPast.append(request_exe_time)
        self.reqEstimation = max(self.reqPast)

    def update_gci_values(self, gc_execution_time):
        # update request history
        if len(self.processed_requests_history) == self.HISTORY_SIZE:
            del self.processed_requests_history[0]
        self.processed_requests_history.append(self.processed_requests)
        self.processed_requests = 0
        # update Check Heap value
        self.check_heap = min(self.processed_requests_history)

        # update GC execution history
        if len(self.gcPast) == self.HISTORY_SIZE:
            del self.gcPast[0]
        self.gcPast.append(gc_execution_time)
        # update estimated gc execution time
        self.estimated_gc_exec_time = max(self.gcPast)