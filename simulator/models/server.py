from log import get_logger
import simpy
import numpy

class Server(object):

    def __init__(self, env, id, conf, log_path):
        self.env = env
        self.id = id
        self.sleep = float(conf['sleep_time'])
        
        self.queue = simpy.Store(env)               # the queue of requests

        self.logger = get_logger(log_path + "/server.log", "SERVER")

        self.processed_requests = 0
        self.times_interrupted = 0
        self.action = env.process(self.run())

    def run(self):
        while True:
            available_time = self.get_next_availability_time()
            available_until = self.env.now + available_time
            while self.env.now < availability_until:
                if len(self.queue.items) > 0:                     # check if there is any request to be processed
                    request = yield self.queue.get()              # get a request from store
                    yield self.env.process(self.process_request(request))
            
                yield self.env.timeout(self.sleep)
            
            unavailable_time = self.get_next_unavailable_time()
            unavailable_until = self.env.now + unavailable_time
            while self.env.now < unavailable_until:
                requests = self.queue.items
                for request in requests:
                    request.interrupted(unavailable_time)
                    
                yield self.env.timeout(self.sleep)

            
    def get_next_available_time(self):
        # for now we will use normal distribution
        average_availability_time = 10  # define by experimental results!
        standard_desviation = 3         # define by experimental results!
        return numpy.random.normal(average_availability_time, standard_desviation)
    
    def get_next_unavailable_time(self):
        # for now we will use chi-square or BETA distribution
        return 0.5

    def process_request(self, request):
        yield self.env.process(request.run(self.env, self.heap))
        yield self.env.process(request.load_balancer.sucess_request(request))
        self.processed_requests += 1

    def request_arrived(self, request):
        request.arrived_at(self.env.now)
        yield self.queue.put(request)   # put the request at the end of the queue


class ServerWithGCI(Server):

    def __init__(self, env, id, conf, log_path=None):
        super().__init__(env, id, conf, log_path)

    def process_request(self, request):
        before = self.env.now
        yield self.env.process(super().process_request(request))
        after = self.env.now
        self.gci.requestFinished(before - after)

    def request_arrived(self, request):
        yield self.env.process(self.gci.before(request))
