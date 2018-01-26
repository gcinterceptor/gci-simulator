

class GCI(object):

    def __init__(self, conf, env, server):
        self.check_heap = int(conf['check_heap'])
        self.threshold = float(conf['threshold'])
        self.sleep_time = float(conf['sleep_time'])
        self.estimated_gc_exec_time = float(conf['initial_gc_exec_time'])

        self.env = env
        self.server = server
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

        self._time_shedding = 0

    def before(self, request):
        yield self.env.process(self.check())

        if self.is_shedding:
            yield self.env.process(request.load_balancer.shed_request(request, self.server, self._time_shedding))

        else:
            request.arrived_at(self.env.now)
            self.requests_to_process += 1
            self.server.requests_arrived += 1
            yield self.env.process(self.server.process_request(request))

    def check(self):
        if self.processed_requests >= self.check_heap:
            if self.server.heap.level >= self.threshold and not self.is_shedding:
                    self.is_shedding = True
                    self.env.process(self.run_gc())

        yield self.env.timeout(self.sleep_time)

    def run_gc(self):
        self._time_shedding = self.estimated_shed_time()
        # wait for the queue to get empty.
        yield self.env.process(self.check_request_queue())

        # "run GC"
        gc_start_time = self.env.now
        yield self.env.process(self.server.run_gc_collect())
        gc_end_time = self.env.now

        # leave server
        self.is_shedding = False
        self._time_shedding = 0

        gc_exec_time = gc_end_time - gc_start_time
        self.update_gci_values(gc_exec_time)

        self.times_performed += 1
        self.requests_to_process = 0
        self.processed_requests = 0

    def check_request_queue(self):
        while self.requests_to_process > self.processed_requests:
            yield self.env.timeout(self.sleep_time)
        yield self.env.timeout(self.sleep_time)

    def estimated_shed_time(self):
        return (self.requests_to_process - self.processed_requests) * self.reqEstimation + self.estimated_gc_exec_time

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