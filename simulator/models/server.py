from log import get_logger
from .distributions import Distribution
import simpy

class Server(object):

    def __init__(self, env, id, conf, log_path, available_avg_time, unavailable_avg_time):
        self.env = env
        self.id = id
        self.sleep = float(conf['sleep_time'])
        
        self.queue = simpy.Store(env)               # the queue of requests
        self.remaining_queue = simpy.Store(env)     # the queue of interrupted requests
        
        if log_path:
            self.logger = get_logger(log_path + "/server.log", "SERVER")
        else:
            self.logger = None
            
        self.available_time_dist = Distribution(conf['available_distribution'], [available_avg_time])
        self.unavailable_time_dist = Distribution(conf['unavailable_distribution'], [unavailable_avg_time])

        self.processed_requests = 0
        self.times_interrupted = 0
        self.action = env.process(self.run())

    def run(self):
        while True:
            available_until = self.env.now + self.get_next_available_time()
            while self.env.now < available_until:
                request = None
                if len(self.remaining_queue.items) > 0:
                    request = yield self.remaining_queue.get()
                elif len(self.queue.items) > 0:                     # check if there is any request to be processed
                    request = yield self.queue.get()                # get a request from store
            
                if request:
                    yield self.env.process(self.process_request(request))
                
                yield self.env.timeout(self.sleep)
            
            unavailable_time = self.get_next_unavailable_time()
            unavailable_until = self.env.now + unavailable_time
            while self.env.now < unavailable_until:
                if len(self.queue.items) > 0:
                    request = yield self.queue.get()
                    yield self.env.process(self.interrupt_request(request, unavailable_time))
                    
                yield self.env.timeout(self.sleep)
            
            self.times_interrupted += 1

    def request_arrived(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The request %d arrived at Server %d" % (self.env.now, request.id, self.id))
        
        yield self.env.process(request.arrived_at())
        yield self.queue.put(request)   # put the request at the end of the queue
    
    def process_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The request %d was processed at Server %d" % (self.env.now, request.id, self.id))
        
        yield self.env.process(request.run(self.id))
        yield self.env.process(request.load_balancer.success_request(request))
        self.processed_requests += 1
        
    def interrupt_request(self, request, unavailable_time):
        if self.logger:
            self.logger.info(" At %.3f, The request %d was interrupted at Server %d" % (self.env.now, request.id, self.id))
        
        yield self.remaining_queue.put(request)
        yield self.env.process(request.interrupted_at(unavailable_time))
        
    def get_next_available_time(self):
        available_time = self.available_time_dist.get_next_value()
        
        if self.logger:
            self.logger.info(" At %.3f, The server %d will be available for %.3f" % (self.env.now, self.id, available_time))
            
        return available_time
    
    def get_next_unavailable_time(self):
        unavailable_time = self.unavailable_time_dist.get_next_value()
        
        if self.logger:
            self.logger.info(" At %.3f, The server %d will be UNavailable for %.3f" % (self.env.now, self.id, unavailable_time))
            
        return unavailable_time
        
class ServerWithGCI(Server):

    def __init__(self, env, id, conf, log_path, avg_available_time, avg_unavailable_time):
        super().__init__(env, id, conf, log_path, avg_available_time, avg_unavailable_time)

    def interrupt_request(self, request, unavailable_time):
        if self.logger:
            self.logger.info(" At %.3f, The request %d was interrupted at Server %d" % (self.env.now, request.id, self.id))
        # Note that we are assuming that the GCI is giving a good hint at unvailable server time
        yield self.env.process(request.load_balancer.shed_request(request, self, unavailable_time))

