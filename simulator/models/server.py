from log import get_logger
from .distributions import Distribution
import simpy

class Server(object):

    def __init__(self, env, id, conf, log_path, available_avg_time, unavailable_avg_time, seed):
        self.env = env
        self.id = id
        
        if log_path:
            self.logger = get_logger(log_path + "/server.log", "SERVER")
        else:
            self.logger = None
        
        self.is_shedding = False
            
        self.available_time_dist = Distribution(conf['available_distribution'], [available_avg_time], seed)
        self.unavailable_time_dist = Distribution(conf['unavailable_distribution'], [unavailable_avg_time], seed)

        self.processed_requests = 0
        self.times_interrupted = 0
        
        self.env.process(self.run())

    def run(self):
        while True:
            available_time = self.get_next_available_time()
            available_until = self.env.now + available_time
            self.is_shedding = False
            yield self.env.timeout(available_time)
            
            unavailable_time = self.get_next_unavailable_time()
            unavailable_until = self.env.now + unavailable_time
            self.is_shedding = True
            self.times_interrupted += 1
            yield self.env.timeout(unavailable_time)
            
    def request_arrived(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The request %d arrived at Server %d" % (self.env.now, request.id, self.id))
        
        request.arrived_at()
        if(self.is_shedding):
            self.interrupt_request(request)
        else:
            self.env.process(self.process_request(request))
    
    def process_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The request %d was processed at Server %d" % (self.env.now, request.id, self.id))
        
        request.processed_at(self.id)
        yield self.env.timeout(request.cpu_time)
        
        self.processed_requests += 1
        self.env.process(request.load_balancer.success_request(request))
        
    def interrupt_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The request %d was interrupted at Server %d" % (self.env.now, request.id, self.id))

        self.env.process(request.load_balancer.shed_request(request, self))
        
    def get_next_available_time(self):
        available_time = self.available_time_dist.get_next_value()
        
        if self.logger:
            self.logger.info(" At %.3f, The server %d will be available for %.3f" % (self.env.now, self.id, available_time))
            
        return available_time
    
    def get_next_unavailable_time(self):
        unavailable_time = self.unavailable_time_dist.get_next_value()
        
        if self.logger:
            self.logger.info(" At %.3f, The server %d will be unavailable for %.3f" % (self.env.now, self.id, unavailable_time))
            
        return unavailable_time
        
