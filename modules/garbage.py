import simpy

class GC(object):

    def __init__(self, env, heap, threshold=0.9, sleep=2):
        self.env = env
        self.heap = heap
        self.threshold = threshold
        self.sleep = sleep
        self.collect_exe = None

    def run(self, server):
        try:
            while True:
                if self.heap.level >= self.threshold:
                    self.collect_exe = self.env.process(self.collect(server))

                yield self.env.timeout(self.sleep)

        except simpy.Interrupt:
            yield self.env.timeout(self.sleep)  # wait for...

    def collect(self, server):
        print("At %.3f, GC GC is running. We have %.3f of trash" % (self.env.now, self.heap.level))

        #server.action.interrupt()

        while self.heap.level > 0:                                          # while threshold is empty...
            trash = self.heap.level                                         # keeps the amount of trash
            yield self.env.timeout(self.gc_execution_time_by_trash(trash))  # run the time of discarting
            yield self.heap.get(trash)                                      # discards the trash

        print("At %.3f, GC GC finish his job. Now we have %.3f of trash" % (self.env.now, self.heap.level))

        server.action = self.env.process(server.run())
        server.gc_times_performed += 1

    def gc_execution_time_by_trash(self, trash):
        """ implement the way to calculate the execution time of Garbage Collector """
        return trash


class GCI(object):

    def __init__(self, env,  threshold=0.7, check_heap=2, initial_gc_exec_time=0.200):
        self.env = env
        self.threshold = threshold
        self.check_heap = check_heap
        self.gc_exec_time = initial_gc_exec_time

        self.shed_requests = False

        self.processed_requests_history = list()
        self.gc_execution_history = list()
        self.history_size = 5

        self.sleep = 0.02

    def run(self, server):
        while True:
            if server.processed_requests >= self.check_heap:
                print("At %.3f, GCI check the heap and it is at %.3f" % (self.env.now, server.heap.level))

                if server.heap.level >= self.threshold:
                    # create an event that will set shed as true and available as false
                    self.shed_requests = True

                    # wait for the queue to get empty.
                    yield self.env.process(self.check_request_queue(server))

                    # run GC
                    gc_start_time = self.env.now
                    yield self.env.process(server.gc.collect(server))
                    gc_end_time = self.env.now

                    # leave server
                    self.update_gci_values(gc_end_time - gc_start_time, server)
                    self.shed_requests = False

            yield self.env.timeout(self.sleep)

    def check_request_queue(self, server):
        while len(server.queue.items) > 0:
            yield self.env.timeout(self.sleep)

    def estimated_shed_time(self, server):
        return self.env.now + self.estimated_request_execution_time(server) + self.estimated_gc_execution_time(server)

    def estimated_request_execution_time(self, server):
        # implement the way to get the max value of the last five time of requests execution time
        # history of lasts requests time executions
        return self.gc_exec_time + len(server.queue.items) * 0.035

    def estimated_gc_execution_time(self, server):
        # implement the way to get the max value of the last five time of requests execution time
        # history of lasts requests time executions
        return self.gc_exec_time + server.heap.level + len(server.queue.items) * 0.01

    def update_gci_values(self, gc_execution_time, server):
        # update request history
        if (len(self.processed_requests_history) == self.history_size):
            del self.processed_requests_history[0]
        self.processed_requests_history.append(server.processed_requests)

        # update Check Heap value
        self.check_heap = min(self.processed_requests_history)

        # update GC execution history
        if (len(self.gc_execution_history) == self.history_size):
            del self.gc_execution_history[0]
        self.gc_execution_history.append(gc_execution_time)

        # update estimated gc execution time
        self.gc_exec_time = max(self.gc_execution_history)