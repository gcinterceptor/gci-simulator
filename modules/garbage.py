import simpy

class GC(object):

    def __init__(self, env, server, threshold=0.9, sleep=0.02):
        self.env = env
        self.server = server
        self.heap = server.heap
        self.threshold = threshold
        self.sleep = sleep
        self.collect_exe = False
        self.times_performed = 0
        self.gc_process = self.env.process(self.run())

    def run(self):
        while True:
            if self.heap.level >= self.threshold and not self.collect_exe:
                yield self.env.process(self.collect())
                self.times_performed += 1

            yield self.env.timeout(self.sleep)  # wait for...

    def collect(self):
        print("At %.3f, GC is running. We have %.3f of trash" % (self.env.now, self.heap.level))
        self.collect_exe = True
        self.server.action.interrupt()
        while self.heap.level > 0:                                          # while threshold is empty...
            trash = self.heap.level                                         # keeps the amount of trash
            yield self.env.timeout(self.gc_execution_time_by_trash(trash))  # run the time of discarting
            yield self.heap.get(trash)                                      # discards the trash

        self.collect_exe = False
        print("At %.3f, GC finish his job. Now we have %.3f of trash" % (self.env.now, self.heap.level))
        self.server.action = self.env.process(self.server.run())

    def gc_execution_time_by_trash(self, trash):
        """ implement the way to calculate the execution time of Garbage Collector """
        return trash


class GCI(object):

    def __init__(self, env, server,  threshold=0.7, check_heap=2, initial_gc_exec_time=0, sleep=0.00002):
        self.env = env
        self.server = server
        self.threshold = threshold
        self.check_heap = check_heap
        self.gc_exec_time = initial_gc_exec_time
        self.sleep = sleep

        self.shed_requests = False
        self.processed_requests_history = list()
        self.gc_execution_history = list()
        self.history_size = 5
        self.times_performed = 0

    def check(self):
        if self.server.processed_requests >= self.check_heap:
            if self.server.heap.level >= self.threshold:
                    yield self.env.process(self.run_gc())

    def run_gc(self):
        print("At %.3f, GCI check the heap and it is at %.3f" % (self.env.now, self.server.heap.level))
        # create an event that will set shed as true and available as false
        self.shed_requests = True

        # wait for the queue to get empty.
        yield self.env.process(self.check_request_queue())

        # run GC
        gc_start_time = self.env.now
        if self.server.gc.collect_exe:
            while self.server.gc.collect_exe:
                yield self.env.timeout(self.sleep)
        else:
            yield self.env.process(self.server.gc.collect())
        gc_end_time = self.env.now

        # leave server
        self.update_gci_values(gc_end_time - gc_start_time)
        self.shed_requests = False
        self.times_performed += 1
        print("At %.3f, GCI finish his job" % (self.env.now))

    def check_request_queue(self):
        while len(self.server.queue.items) > 0:
            yield self.env.timeout(self.sleep)

    def estimated_shed_time(self):
        return self.estimated_request_execution_time() + self.estimated_gc_execution_time()

    def estimated_request_execution_time(self):
        # implement the way to get the max value of the last five time of requests execution time
        # history of lasts requests time executions
        return self.gc_exec_time + len(self.server.queue.items) * 0.001

    def estimated_gc_execution_time(self):
        # implement the way to get the max value of the last five time of requests execution time
        # history of lasts requests time executions
        return self.gc_exec_time + self.server.heap.level + len(self.server.queue.items) * 0.1

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
        self.gc_exec_time = max(self.gc_execution_history)