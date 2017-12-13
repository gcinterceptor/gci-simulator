from log import get_logger
import simpy
import numpy

class Server(object):

    def __init__(self, env, id, conf, log_path):
        self.env = env
        self.id = id
        self.sleep = float(conf['sleep_time'])
        
        self.queue = simpy.Store(env)               # the queue of requests
        
        if log_path:
            self.logger = get_logger(log_path + "/server.log", "SERVER")
        else:
            self.logger = None

        self.processed_requests = 0
        self.times_interrupted = 0
        self.action = env.process(self.run())

    def run(self):
        while True:
            available_time = self.get_next_available_time()
            available_until = self.env.now + available_time
            
            while self.env.now < available_until:
                if len(self.queue.items) > 0:                     # check if there is any request to be processed
                    request = yield self.queue.get()              # get a request from store
                    self.process_request(request)
            
                yield self.env.timeout(self.sleep)
            
            unavailable_time = self.get_next_unavailable_time()
            unavailable_until = self.env.now + unavailable_time
            
            while self.env.now < unavailable_until:
                requests = self.queue.items
                for request in requests:
                    self.interrupt_request(request, unavailable_time)
                    
                yield self.env.timeout(self.sleep)

            
    def get_next_available_time(self):
        # for now we will use normal distribution
        #average_availability_time = 10  # defined by experimental results!
        #standard_desviation = 3         # defined by experimental results!
        #return numpy.random.normal(average_availability_time, standard_desviation)
        return 0.9
    
    def get_next_unavailable_time(self):
        # for now we will use chi-square or BETA distribution
        return 0.1

    def process_request(self, request):
        yield self.env.process(request.run(self.env))
        yield self.env.process(request.load_balancer.success_request(request))
        self.processed_requests += 1
        
    def interrupt_request(self, request, unavailable_time):
        request.interrupted_at(unavailable_time)

    def request_arrived(self, request):
        request.arrived_at(self.env.now)
        yield self.queue.put(request)   # put the request at the end of the queue


class ServerWithGCI(Server):

    def __init__(self, env, id, conf, log_path=None):
        super().__init__(env, id, conf, log_path)

    def interrupt_request(self, request, unavailable_time):
        # Note that we are assuming that the GCI is giving a good hint at unvailable server time
        yield self.env.process(request.load_balancer.shed_request(request, self.server, unavailable_time))

