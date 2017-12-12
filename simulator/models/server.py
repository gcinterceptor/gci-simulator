from log import get_logger
import simpy
import numpy

class Server(object):

    def __init__(self, env, id, conf, gc_conf, log_path):
        self.env = env
        self.id = id
        self.sleep = float(conf['sleep_time'])

        self.logger = get_logger(log_path + "/server.log", "SERVER")

        self.processed_requests = 0
        self.times_interrupted = 0
        self.action = env.process(self.run())

    def run(self):
        try:
            while True:
                if len(self.interrupted_queue.items) > 0:
                    request = yield self.interrupted_queue.get()    # get a request from store
                    if request.done:
                        yield self.heap.get(request.memory)         # remove trash that shouldn't be added...
                    yield self.env.process(self.process_request(request))

                elif len(self.queue.items) > 0:                     # check if there is any request to be processed
                    request = yield self.queue.get()                # get a request from store
                    yield self.env.process(self.process_request(request))

                yield self.env.timeout(self.sleep)                  # wait for...

        except simpy.Interrupt:
            self.logger.info(" At %.3f, Server was interrupted" % (self.env.now))
            self.times_interrupted += 1
            yield self.interrupted_queue.put(request)
            
    def get_next_availability_time(self):
        average_availability_time = 10  # define by experimental results!
        standard_desviation = 3         # define by experimental results!
        return self.env.now + numpy.random.normal(average_availability_time, standard_desviation)
    
    def get_next_unavailability_time(self):
        return 1

    def process_request(self, request):
        yield self.env.process(request.run(self.env, self.heap))
        yield self.env.process(request.load_balancer.sucess_request(request))
        self.processed_requests += 1

    def request_arrived(self, request):
        request.arrived_at(self.env.now)
        yield self.queue.put(request)   # put the request at the end of the queue

