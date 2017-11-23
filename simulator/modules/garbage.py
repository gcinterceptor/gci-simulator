from utils import getLogger

class GC(object):

    def __init__(self, env, server, conf, log_path):
        self.env = env
        self.server = server
        self.heap = server.heap
        self.threshold = float(conf['threshold'])
        self.sleep_time = float(conf['sleep_time'])

        self.is_gcing = False
        self.times_performed = 0
        self.collects_performed = 0
        self.gc_exec_time_sum = 0
        self.gc_process = self.env.process(self.run())

        self.logger = getLogger(log_path + "/gc.log", "GC")

    def run(self):
        while True:
            if self.heap.level >= self.threshold and not self.is_gcing:
                yield self.env.process(self.collect())
                self.times_performed += 1

            yield self.env.timeout(self.sleep_time)  # wait for...

    def collect(self):
        self.logger.info(" At %.3f, GC is running. We have %.3f of trash" % (self.env.now, self.heap.level))
        self.is_gcing = True
        self.server.action.interrupt()

        gc_start_time = self.env.now
        while self.heap.level > 0:                                          # while threshold is empty...
            trash = self.heap.level                                         # keeps the amount of trash
            yield self.env.timeout(self.gc_execution_time_by_trash(trash))  # run the time of discarting
            yield self.heap.get(trash)                                      # discards the trash
        self.gc_exec_time_sum += (self.env.now - gc_start_time)

        self.is_gcing = False
        self.logger.info(" At %.3f, GC finish his job. Now we have %.3f of trash\n" % (self.env.now, self.heap.level))
        self.server.action = self.env.process(self.server.run())
        self.collects_performed += 1

    def gc_execution_time_by_trash(self, trash):
        # TODO(David) implement the way to calculate the execution time of Garbage Collector
        return trash

class GCI(object):

    def __init__(self, env, server,  conf, log_path):
        self.env = env
        self.server = server
        self.check_heap = int(conf['check_heap'])
        self.threshold = float(conf['threshold'])
        self.sleep_time = float(conf['sleep_time'])
        self.estimated_gc_exec_time = float(conf['initial_estimated_gc_exec_time'])

        self.is_shedding = False
        self.processed_requests_history = list()
        self.gc_execution_history = list()
        self.history_size = 5
        self.times_performed = 0

        self.logger = getLogger(log_path + "/gci.log", "GCI")

        self._time_shedding = 0

    def intercept(self, request):
        self.env.process(self.check())
        self.env.timeout(self.sleep_time)

        if self.is_shedding:
            self.logger.info(" At %.3f, shedding request" % self.env.now)
            yield self.env.process(request.load_balancer.shed_request(request, self.server, self._time_shedding))

        else:
            yield self.server.queue.put(request)  # put the request at the end of the queue
            yield self.env.process(request.load_balancer.successfully_sent(request))

    def check(self):
        if self.server.processed_requests >= self.check_heap:
            if self.server.heap.level >= self.threshold and not self.is_shedding:
                    yield self.env.process(self.run_gc())

    def run_gc(self):
        self.logger.info(" At %.3f, GCI check the heap and it is at %.3f" % (self.env.now, self.server.heap.level))
        # create an event that will set shed as true and available as false

        self.is_shedding = True
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
        self.logger.info(" At %.3f, GCI finish his job and GC takes %.3f seconds to execute\n" % (self.env.now, gc_exec_time))

    def check_request_queue(self):
        while len(self.server.queue.items) > 0:
            yield self.env.timeout(self.sleep_time)

    def estimated_shed_time(self):
        return self.estimated_requests_execution_time() + self.estimated_gc_execution_time()

    def estimated_requests_execution_time(self):
        # TODO(David) implement the logic to estimate time to process one request
        return len(self.server.queue.items) * 0.001

    def estimated_gc_execution_time(self):
        return self.estimated_gc_exec_time

    def update_gci_values(self, gc_execution_time):
        # update request history
        if (len(self.processed_requests_history) == self.history_size):
            del self.processed_requests_history[0]
        self.processed_requests_history.append(self.server.processed_requests)

        # update Check Heap value
        self.check_heap = min(self.processed_requests_history)

        # update GC execution history
        if (len(self.gc_execution_history) == self.history_size):
            del self.gc_execution_history[0]
        self.gc_execution_history.append(gc_execution_time)

        # update estimated gc execution time
        self.estimated_gc_exec_time = max(self.gc_execution_history)